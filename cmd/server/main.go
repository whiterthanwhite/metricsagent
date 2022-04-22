package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/whiterthanwhite/metricsagent/internal/handlers"
	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
	"github.com/whiterthanwhite/metricsagent/internal/settings"
	"github.com/whiterthanwhite/metricsagent/internal/storage"
)

var (
	ServerSettings    = settings.GetSysSettings()
	flagAddress       *string
	flagRestore       *bool
	flagStoreInterval *time.Duration
	flagStoreFile     *string
)

func init() {
	flagAddress = flag.String("a", settings.DefaultAddress, "")
	flagRestore = flag.Bool("r", settings.DefaultRestore, "")
	flagStoreInterval = flag.Duration("i", settings.DefaultStoreInterval, "")
	flagStoreFile = flag.String("f", settings.DefaultStoreFile, "")
}

func startSaveMetricsOnFile(serverMetrics map[string]metrics.Metrics) {
	if ServerSettings.StoreFile == "" {
		return
	}

	saveTicker := time.NewTicker(ServerSettings.StoreInterval)
	defer saveTicker.Stop()
	for {
		<-saveTicker.C
		saveMetricsOnFile(serverMetrics)
	}
}

func saveMetricsOnFile(serverMetrics map[string]metrics.Metrics) {
	if ServerSettings.StoreFile == "" {
		return
	}
	producer, err := storage.NewProducer(ServerSettings.StoreFile)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()
	if err := producer.WriteMetrics(serverMetrics); err != nil {
		log.Fatal(err)
	}
}

func restoreMetricsFromFile() map[string]metrics.Metrics {
	var serverMetrics map[string]metrics.Metrics = nil
	if ServerSettings.Restore && ServerSettings.StoreFile != "" {
		consumer, err := storage.NewConsumer(ServerSettings.StoreFile)
		if err != nil {
			log.Fatal(err)
		}
		defer consumer.Close()
		serverMetrics, err = consumer.ReadMetrics()
		if err != nil {
			log.Println(err)
		}
		log.Println(serverMetrics)
		log.Println("restored")
	}
	if serverMetrics == nil {
		serverMetrics = metrics.GetAllNewMetrics()
	}
	return serverMetrics
}

func main() {
	log.Println("Server start")

	flag.Parse()
	log.Println(ServerSettings.Address, *flagAddress)
	if ServerSettings.Address == settings.DefaultAddress {
		ServerSettings.Address = *flagAddress
	}
	log.Println(ServerSettings.Restore, *flagRestore)
	if ServerSettings.Restore == settings.DefaultRestore {
		ServerSettings.Restore = ServerSettings.Restore || *flagRestore
	}
	log.Println(ServerSettings.StoreInterval, *flagStoreInterval)
	if ServerSettings.StoreInterval == settings.DefaultStoreInterval {
		ServerSettings.StoreInterval = *flagStoreInterval
	}
	log.Println(ServerSettings.StoreFile, *flagStoreFile)
	if ServerSettings.StoreFile == settings.DefaultStoreFile {
		ServerSettings.StoreFile = *flagStoreFile
	}
	log.Println(ServerSettings)

	newServerMetrics := restoreMetricsFromFile()
	oldServerMetrics := metrics.GetAllMetrics()
	defer saveMetricsOnFile(newServerMetrics)
	go startSaveMetricsOnFile(newServerMetrics)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", handlers.GetAllMetricsFromFile(oldServerMetrics))
		r.Route("/update", func(r chi.Router) {
			r.Post("/", handlers.UpdateMetricOnServer(newServerMetrics))
			r.Post("/{metricType}/{metricName}/{metricValue}",
				handlers.UpdateMetricHandler(oldServerMetrics, newServerMetrics))
		})
		r.Route("/value", func(r chi.Router) {
			r.Post("/", handlers.GetMetricFromServer(newServerMetrics))
			r.Get("/{metricType}/{metricName}",
				handlers.GetMetricValueFromServer(oldServerMetrics))
		})
		// r.Post("/", handlers.GetAllMetricsFromServer(serverMetrics))
	})

	port := fmt.Sprintf(":%v", strings.Split(ServerSettings.Address, ":")[1])
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal(err)
	}
	log.Println("Server Stop")
}
