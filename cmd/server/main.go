package main

import (
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
	ServerSettings = settings.GetSysSettings()
)

func startSaveMetricsOnFile(serverMetrics map[string]metrics.Metrics) {
	if ServerSettings.StoreFile == "" {
		return
	}

	saveTicker := time.NewTicker(ServerSettings.StoreInterval)
	for {
		select {
		case <-saveTicker.C:
			saveMetricsOnFile(serverMetrics)
		}
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
	producer.WriteMetrics(serverMetrics)
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
	}
	if serverMetrics == nil {
		serverMetrics = metrics.GetAllNewMetrics()
	}
	return serverMetrics
}

func main() {
	newServerMetrics := restoreMetricsFromFile()
	oldServerMetrics := metrics.GetAllMetrics()
	defer saveMetricsOnFile(newServerMetrics)
	go saveMetricsOnFile(newServerMetrics)

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
	log.Println(ServerSettings)
	log.Fatal(http.ListenAndServe(port, r))
}
