package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/whiterthanwhite/metricsagent/internal/handlers"
	"github.com/whiterthanwhite/metricsagent/internal/metricdb"
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
	flagHashKey       *string
	flagDBAddress     *string
)

func init() {
	flagAddress = flag.String("a", settings.DefaultAddress, "")
	flagRestore = flag.Bool("r", settings.DefaultRestore, "")
	flagStoreInterval = flag.Duration("i", settings.DefaultStoreInterval, "")
	flagStoreFile = flag.String("f", settings.DefaultStoreFile, "")
	flagHashKey = flag.String("k", settings.DefaultHashKey, "")
	flagDBAddress = flag.String("d", settings.DefaultHashKey, "")
}

func startSaveMetricsOnFile(serverMetrics map[string]metrics.Metrics) {
	if ServerSettings.StoreFile == "" {
		return
	}

	saveTicker := time.NewTicker(ServerSettings.StoreInterval)
	defer saveTicker.Stop()
	for {
		<-saveTicker.C
		storage.SaveMetricsOnFile(serverMetrics, ServerSettings)
	}
}

func main() {
	log.Println("Server start")

	flag.Parse()
	if ServerSettings.Address == settings.DefaultAddress {
		ServerSettings.Address = *flagAddress
	}
	if ServerSettings.Restore == settings.DefaultRestore {
		ServerSettings.Restore = ServerSettings.Restore || *flagRestore
	}
	if ServerSettings.StoreInterval == settings.DefaultStoreInterval {
		ServerSettings.StoreInterval = *flagStoreInterval
	}
	if ServerSettings.StoreFile == settings.DefaultStoreFile {
		ServerSettings.StoreFile = *flagStoreFile
	}
	if ServerSettings.Key == settings.DefaultHashKey {
		ServerSettings.Key = *flagHashKey
	}
	if ServerSettings.MetricDBAdress == settings.DefaultDBAddress {
		ServerSettings.MetricDBAdress = *flagDBAddress
	}
	log.Println(ServerSettings)

	newServerMetrics := storage.RestoreMetricsFromFile(ServerSettings)
	oldServerMetrics := metrics.GetAllMetrics()
	defer storage.SaveMetricsOnFile(newServerMetrics, ServerSettings)
	go startSaveMetricsOnFile(newServerMetrics)

	// postgresURLString := "postgres://localhost:5432/metricsagentdb"
	mdb := metricdb.CreateDBConnnect(context.Background(), ServerSettings.MetricDBAdress)
	defer mdb.DBClose()

	r := chi.NewRouter()

	// TODO: Add middleware
	r.Route("/", func(r chi.Router) {
		r.Get("/", handlers.GetAllMetricsFromFile(oldServerMetrics, newServerMetrics))
		r.Route("/update", func(r chi.Router) {
			r.Post("/", handlers.UpdateMetricOnServer(newServerMetrics, ServerSettings))
			r.Post("/{metricType}/{metricName}/{metricValue}",
				handlers.UpdateMetricHandler(oldServerMetrics, newServerMetrics))
		})
		r.Route("/value", func(r chi.Router) {
			r.Post("/", handlers.GetMetricFromServer(newServerMetrics, ServerSettings))
			r.Get("/{metricType}/{metricName}",
				handlers.GetMetricValueFromServer(oldServerMetrics))
		})
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", handlers.CheckDatabaseConn(mdb))
		})
		// r.Post("/", handlers.GetAllMetricsFromServer(serverMetrics))
	})

	port := fmt.Sprintf(":%v", strings.Split(ServerSettings.Address, ":")[1])
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal(err)
	}
	log.Println("Server Stop")
}
