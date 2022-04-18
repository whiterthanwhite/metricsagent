package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/whiterthanwhite/metricsagent/internal/handlers"
	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
	"github.com/whiterthanwhite/metricsagent/internal/storage"
)

func main() {
	metricFile := storage.OpenMetricFileCSV()
	defer metricFile.Close()

	// Set exist metrics or create empty
	addedMetrics := storage.GetMetricsFromFile(metricFile)
	if len(addedMetrics) == 0 {
		addedMetrics = metrics.GetAllMetrics()
	}

	r := chi.NewRouter()

	/*
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
		})
	*/

	serverMetrics := make([]metrics.NewMetric, 0)
	r.Route("/", func(r chi.Router) {
		r.Post("/", handlers.GetAllMetricsFromServer(serverMetrics))
		r.Post("/update/", handlers.UpdateMetricOnServer(&serverMetrics)) // update metric
		r.Post("/value/", handlers.GetMetricFromServer(serverMetrics))
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
