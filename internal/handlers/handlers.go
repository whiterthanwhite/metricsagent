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
		/*
				if r.Method != http.MethodPost {
					http.Error(rw, "status method is not allowed", http.StatusMethodNotAllowed)
					return
				}


			headerContentType := r.Header.Get("Content-Type")
			if headerContentType != "text/plain" {
				http.Error(rw, "Unsupported Media Type", 415)
				return
			}
		*/

		// Parse URL
		metricURL := r.URL
		metricURIValues := strings.Split(metricURL.RequestURI(), "/")
		if len(metricURIValues) < 5 {
			http.Error(rw, "", http.StatusNotFound)
			return
		}

		mType := metricURIValues[2]
		if !metrics.IsMetricTypeExist(mType) {
			http.Error(rw, "", http.StatusNotImplemented)
			return
		}

		m, ok := getMetricFromValues(metricURIValues)
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

func getMetricFromValues(sendedValues []string) (metrics.Metric, bool) {
	mType := sendedValues[2]
	metricName := sendedValues[3]
	metricValue := sendedValues[4]

	// debug >>
	if len(addedMetrics) == 0 {
		addedMetrics = make(map[string]metrics.Metric)
	}
	// debug <<
	m, ok := addedMetrics[metricName]
	if !ok {
		m = metrics.GetMetric(metricName, mType)
		addedMetrics[metricName] = m
	}

	value, err := strconv.ParseFloat(metricValue, 64)
	if err != nil {
		fmt.Println("Cannot update metric value")
		return nil, false
	}
	m.UpdateValue(value)

	return m, true
}

func SetMetrics(metrics map[string]metrics.Metric) {
	addedMetrics = metrics
}
