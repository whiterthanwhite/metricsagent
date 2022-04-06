package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/whiterthanwhite/metricsagent/internal/handlers"
	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
	"github.com/whiterthanwhite/metricsagent/internal/storage"
)

func main() {
	var metricFile *os.File = storage.OpenMetricFileCSV()
	defer metricFile.Close()

	// Set exist metrics or create empty
	addedMetrics := storage.GetMetricsFromFile(metricFile)
	if len(addedMetrics) == 0 {
		addedMetrics = metrics.GetAllMetrics()
	}
	handlers.SetMetrics(addedMetrics)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", handlers.GetAllMetricsFromFile())
		r.Route("/update", func(r chi.Router) {
			r.Post("/{metricType}/{metricName}/{metricValue}",
				handlers.UpdateMetricHandler(metricFile))
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/{metricType}/{metricName}",
				handlers.GetMetricValueFromServer(metricFile))
		})
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
