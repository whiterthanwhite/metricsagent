package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
	"github.com/whiterthanwhite/metricsagent/internal/storage"
)

var (
	addedMetrics map[string]metrics.Metric
)

func UpdateMetricHandler(f *os.File) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		// Check header
		if r.Method != http.MethodPost {
			http.Error(rw, "status method is not allowed", http.StatusMethodNotAllowed)
			return
		}

		headerContentType := r.Header.Get("Content-Type")
		if headerContentType != "text/plain" {
			http.Error(rw, "Unsupported Media Type", 415)
			return
		}

		// Parse URL
		metricURL := r.URL
		m, ok := parseMetricURI(metricURL.RequestURI())
		if !ok && m == nil {
			http.Error(rw, "Metric wont't found", http.StatusBadRequest)
			return
		}
		addedMetrics[m.GetName()] = m
		storage.WriteMetricsToFile(f, addedMetrics)

		rw.Header().Add("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusOK)
	}
}

func parseMetricURI(metricURI string) (metrics.Metric, bool) {
	var metricURIValues []string = strings.Split(metricURI, "/")
	// var metricType string = metricURIValues[2]
	var metricName string = metricURIValues[3]
	var metricValue string = metricURIValues[4]

	// debug >>
	if len(addedMetrics) == 0 {
		addedMetrics = make(map[string]metrics.Metric)
	}
	// debug <<
	m, ok := addedMetrics[metricName]
	if !ok {
		addedMetrics[metricName] = metrics.GetMetric(metricName)
		m = addedMetrics[metricName]
	}

	value, err := strconv.ParseFloat(metricValue, 64)
	if err != nil {
		fmt.Println("Connot update metric value")
		return nil, false
	}
	m.UpdateValue(value)

	return m, true
}

func SetMetrics(metrics map[string]metrics.Metric) {
	addedMetrics = metrics
}
