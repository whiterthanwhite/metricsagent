package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/whiterthanwhite/metricsagent/internal/handlers"
	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
	"github.com/whiterthanwhite/metricsagent/internal/storage"
)

func getTempServerMetrics() map[string]metrics.NewMetric {
	ms := make(map[string]metrics.NewMetric)
	ms["PollCount"] = metrics.NewMetric{ID: "PollCount", MType: metrics.CounterType}
	ms["Alloc"] = metrics.NewMetric{ID: "Alloc", MType: metrics.GaugeType}
	ms["BuckHashSys"] = metrics.NewMetric{ID: "BuckHashSys", MType: metrics.GaugeType}
	ms["Frees"] = metrics.NewMetric{ID: "Frees", MType: metrics.GaugeType}
	ms["GCCPUFraction"] = metrics.NewMetric{ID: "GCCPUFraction", MType: metrics.GaugeType}
	ms["GCSys"] = metrics.NewMetric{ID: "GCSys", MType: metrics.GaugeType}
	ms["HeapAlloc"] = metrics.NewMetric{ID: "HeapAlloc", MType: metrics.GaugeType}
	ms["HeapIdle"] = metrics.NewMetric{ID: "HeapIdle", MType: metrics.GaugeType}
	ms["HeapInuse"] = metrics.NewMetric{ID: "HeapInuse", MType: metrics.GaugeType}
	ms["HeapObjects"] = metrics.NewMetric{ID: "HeapObjects", MType: metrics.GaugeType}
	ms["HeapReleased"] = metrics.NewMetric{ID: "HeapReleased", MType: metrics.GaugeType}
	ms["HeapSys"] = metrics.NewMetric{ID: "HeapSys", MType: metrics.GaugeType}
	ms["LastGC"] = metrics.NewMetric{ID: "LastGC", MType: metrics.GaugeType}
	ms["Lookups"] = metrics.NewMetric{ID: "Lookups", MType: metrics.GaugeType}
	ms["MCacheInuse"] = metrics.NewMetric{ID: "MCacheInuse", MType: metrics.GaugeType}
	ms["MCacheSys"] = metrics.NewMetric{ID: "MCacheSys", MType: metrics.GaugeType}
	ms["MSpanInuse"] = metrics.NewMetric{ID: "MSpanInuse", MType: metrics.GaugeType}
	ms["MSpanSys"] = metrics.NewMetric{ID: "MSpanSys", MType: metrics.GaugeType}
	ms["Mallocs"] = metrics.NewMetric{ID: "Mallocs", MType: metrics.GaugeType}
	ms["NextGC"] = metrics.NewMetric{ID: "NextGC", MType: metrics.GaugeType}
	ms["NumForcedGC"] = metrics.NewMetric{ID: "NumForcedGC", MType: metrics.GaugeType}
	ms["NumGC"] = metrics.NewMetric{ID: "NumGC", MType: metrics.GaugeType}
	ms["PauseTotalNs"] = metrics.NewMetric{ID: "PauseTotalNs", MType: metrics.GaugeType}
	ms["StackInuse"] = metrics.NewMetric{ID: "StackInuse", MType: metrics.GaugeType}
	ms["StackSys"] = metrics.NewMetric{ID: "StackSys", MType: metrics.GaugeType}
	ms["Sys"] = metrics.NewMetric{ID: "Sys", MType: metrics.GaugeType}
	ms["TotalAlloc"] = metrics.NewMetric{ID: "TotalAlloc", MType: metrics.GaugeType}
	ms["RandomValue"] = metrics.NewMetric{ID: "RandomValue", MType: metrics.GaugeType}
	return ms
}

func main() {
	metricFile := storage.OpenMetricFileCSV()
	defer metricFile.Close()

	// Set exist metrics or create empty
	addedMetrics := storage.GetMetricsFromFile(metricFile)
	if len(addedMetrics) == 0 {
		addedMetrics = metrics.GetAllMetrics()
	}

	r := chi.NewRouter()

	// serverMetrics := metrics.GetAllMetricsSlices()
	tempServerMetrics := getTempServerMetrics()

	r.Route("/", func(r chi.Router) {
		r.Get("/", handlers.GetAllMetricsFromFile(addedMetrics))
		r.Route("/update", func(r chi.Router) {
			r.Post("/{metricType}/{metricName}/{metricValue}",
				handlers.UpdateMetricHandler(metricFile, addedMetrics))
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/{metricType}/{metricName}",
				handlers.GetMetricValueFromServer(metricFile, addedMetrics))
		})
		// r.Post("/", handlers.GetAllMetricsFromServer(serverMetrics))
		// r.Post("/update/", handlers.UpdateMetricOnServer(&serverMetrics))
		// r.Post("/value/", handlers.GetMetricFromServer(&serverMetrics))
		r.Post("/update/", handlers.UpdateMetricOnServerTemp(tempServerMetrics))
		r.Post("/value/", handlers.GetMetricFromServerTemp(tempServerMetrics))
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
