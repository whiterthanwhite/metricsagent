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
	var a int64 = 0
	var b float64 = 0.0
	ms := make(map[string]metrics.NewMetric)
	ms["PollCount"] = metrics.NewMetric{ID: "PollCount", MType: metrics.CounterType, Delta: &a}
	ms["Alloc"] = metrics.NewMetric{ID: "Alloc", MType: metrics.GaugeType, Value: &b}
	ms["BuckHashSys"] = metrics.NewMetric{ID: "BuckHashSys", MType: metrics.GaugeType, Value: &b}
	ms["Frees"] = metrics.NewMetric{ID: "Frees", MType: metrics.GaugeType, Value: &b}
	ms["GCCPUFraction"] = metrics.NewMetric{ID: "GCCPUFraction", MType: metrics.GaugeType, Value: &b}
	ms["GCSys"] = metrics.NewMetric{ID: "GCSys", MType: metrics.GaugeType, Value: &b}
	ms["HeapAlloc"] = metrics.NewMetric{ID: "HeapAlloc", MType: metrics.GaugeType, Value: &b}
	ms["HeapIdle"] = metrics.NewMetric{ID: "HeapIdle", MType: metrics.GaugeType, Value: &b}
	ms["HeapInuse"] = metrics.NewMetric{ID: "HeapInuse", MType: metrics.GaugeType, Value: &b}
	ms["HeapObjects"] = metrics.NewMetric{ID: "HeapObjects", MType: metrics.GaugeType, Value: &b}
	ms["HeapReleased"] = metrics.NewMetric{ID: "HeapReleased", MType: metrics.GaugeType, Value: &b}
	ms["HeapSys"] = metrics.NewMetric{ID: "HeapSys", MType: metrics.GaugeType, Value: &b}
	ms["LastGC"] = metrics.NewMetric{ID: "LastGC", MType: metrics.GaugeType, Value: &b}
	ms["Lookups"] = metrics.NewMetric{ID: "Lookups", MType: metrics.GaugeType, Value: &b}
	ms["MCacheInuse"] = metrics.NewMetric{ID: "MCacheInuse", MType: metrics.GaugeType, Value: &b}
	ms["MCacheSys"] = metrics.NewMetric{ID: "MCacheSys", MType: metrics.GaugeType, Value: &b}
	ms["MSpanInuse"] = metrics.NewMetric{ID: "MSpanInuse", MType: metrics.GaugeType, Value: &b}
	ms["MSpanSys"] = metrics.NewMetric{ID: "MSpanSys", MType: metrics.GaugeType, Value: &b}
	ms["Mallocs"] = metrics.NewMetric{ID: "Mallocs", MType: metrics.GaugeType, Value: &b}
	ms["NextGC"] = metrics.NewMetric{ID: "NextGC", MType: metrics.GaugeType, Value: &b}
	ms["NumForcedGC"] = metrics.NewMetric{ID: "NumForcedGC", MType: metrics.GaugeType, Value: &b}
	ms["NumGC"] = metrics.NewMetric{ID: "NumGC", MType: metrics.GaugeType, Value: &b}
	ms["PauseTotalNs"] = metrics.NewMetric{ID: "PauseTotalNs", MType: metrics.GaugeType, Value: &b}
	ms["StackInuse"] = metrics.NewMetric{ID: "StackInuse", MType: metrics.GaugeType, Value: &b}
	ms["StackSys"] = metrics.NewMetric{ID: "StackSys", MType: metrics.GaugeType, Value: &b}
	ms["Sys"] = metrics.NewMetric{ID: "Sys", MType: metrics.GaugeType, Value: &b}
	ms["TotalAlloc"] = metrics.NewMetric{ID: "TotalAlloc", MType: metrics.GaugeType, Value: &b}
	ms["RandomValue"] = metrics.NewMetric{ID: "RandomValue", MType: metrics.GaugeType, Value: &b}
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
	log.Println(tempServerMetrics)

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
		r.Post("/update/", handlers.UpdateMetricOnServerTemp(&tempServerMetrics))
		r.Post("/value/", handlers.GetMetricFromServerTemp(&tempServerMetrics))
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
