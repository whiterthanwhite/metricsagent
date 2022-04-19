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
