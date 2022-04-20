package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/whiterthanwhite/metricsagent/internal/handlers"
	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
	"github.com/whiterthanwhite/metricsagent/internal/storage"
)

func getTempServerMetrics() map[string]metrics.Metrics {
	var a int64 = 0
	b := 0.0
	ms := make(map[string]metrics.Metrics)
	ms["PollCount"] = metrics.Metrics{ID: "PollCount", MType: metrics.CounterType, Delta: &a}
	ms["Alloc"] = metrics.Metrics{ID: "Alloc", MType: metrics.GaugeType, Value: &b}
	ms["BuckHashSys"] = metrics.Metrics{ID: "BuckHashSys", MType: metrics.GaugeType, Value: &b}
	ms["Frees"] = metrics.Metrics{ID: "Frees", MType: metrics.GaugeType, Value: &b}
	ms["GCCPUFraction"] = metrics.Metrics{ID: "GCCPUFraction", MType: metrics.GaugeType, Value: &b}
	ms["OtherSys"] = metrics.Metrics{ID: "OtherSys", MType: metrics.GaugeType, Value: &b}
	ms["GCSys"] = metrics.Metrics{ID: "GCSys", MType: metrics.GaugeType, Value: &b}
	ms["HeapAlloc"] = metrics.Metrics{ID: "HeapAlloc", MType: metrics.GaugeType, Value: &b}
	ms["HeapIdle"] = metrics.Metrics{ID: "HeapIdle", MType: metrics.GaugeType, Value: &b}
	ms["HeapInuse"] = metrics.Metrics{ID: "HeapInuse", MType: metrics.GaugeType, Value: &b}
	ms["HeapObjects"] = metrics.Metrics{ID: "HeapObjects", MType: metrics.GaugeType, Value: &b}
	ms["HeapReleased"] = metrics.Metrics{ID: "HeapReleased", MType: metrics.GaugeType, Value: &b}
	ms["HeapSys"] = metrics.Metrics{ID: "HeapSys", MType: metrics.GaugeType, Value: &b}
	ms["LastGC"] = metrics.Metrics{ID: "LastGC", MType: metrics.GaugeType, Value: &b}
	ms["Lookups"] = metrics.Metrics{ID: "Lookups", MType: metrics.GaugeType, Value: &b}
	ms["MCacheInuse"] = metrics.Metrics{ID: "MCacheInuse", MType: metrics.GaugeType, Value: &b}
	ms["MCacheSys"] = metrics.Metrics{ID: "MCacheSys", MType: metrics.GaugeType, Value: &b}
	ms["MSpanInuse"] = metrics.Metrics{ID: "MSpanInuse", MType: metrics.GaugeType, Value: &b}
	ms["MSpanSys"] = metrics.Metrics{ID: "MSpanSys", MType: metrics.GaugeType, Value: &b}
	ms["Mallocs"] = metrics.Metrics{ID: "Mallocs", MType: metrics.GaugeType, Value: &b}
	ms["NextGC"] = metrics.Metrics{ID: "NextGC", MType: metrics.GaugeType, Value: &b}
	ms["NumForcedGC"] = metrics.Metrics{ID: "NumForcedGC", MType: metrics.GaugeType, Value: &b}
	ms["NumGC"] = metrics.Metrics{ID: "NumGC", MType: metrics.GaugeType, Value: &b}
	ms["PauseTotalNs"] = metrics.Metrics{ID: "PauseTotalNs", MType: metrics.GaugeType, Value: &b}
	ms["StackInuse"] = metrics.Metrics{ID: "StackInuse", MType: metrics.GaugeType, Value: &b}
	ms["StackSys"] = metrics.Metrics{ID: "StackSys", MType: metrics.GaugeType, Value: &b}
	ms["Sys"] = metrics.Metrics{ID: "Sys", MType: metrics.GaugeType, Value: &b}
	ms["TotalAlloc"] = metrics.Metrics{ID: "TotalAlloc", MType: metrics.GaugeType, Value: &b}
	ms["RandomValue"] = metrics.Metrics{ID: "RandomValue", MType: metrics.GaugeType, Value: &b}
	return ms
}

func main() {
	serverMetrics := getTempServerMetrics()

	metricFile := storage.OpenMetricFileCSV()
	defer metricFile.Close()

	// Set exist metrics or create empty
	addedMetrics := storage.GetMetricsFromFile(metricFile)
	if len(addedMetrics) == 0 {
		addedMetrics = metrics.GetAllMetrics()
	}

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", handlers.GetAllMetricsFromFile(addedMetrics))
		r.Route("/update", func(r chi.Router) {
			r.Post("/", handlers.UpdateMetricOnServer(serverMetrics))
			r.Post("/{metricType}/{metricName}/{metricValue}",
				handlers.UpdateMetricHandler(addedMetrics, serverMetrics))
		})
		r.Route("/value", func(r chi.Router) {
			r.Post("/", handlers.GetMetricFromServer(serverMetrics))
			r.Get("/{metricType}/{metricName}",
				handlers.GetMetricValueFromServer(addedMetrics))
		})
		// r.Post("/", handlers.GetAllMetricsFromServer(serverMetrics))
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
