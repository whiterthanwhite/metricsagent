package main

import (
	"net/http"
	"os"

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

	http.HandleFunc("/update/", handlers.UpdateMetricHandler(metricFile))

	http.ListenAndServe(":8080", nil)
}
