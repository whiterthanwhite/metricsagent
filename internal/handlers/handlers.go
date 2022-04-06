package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
	"github.com/whiterthanwhite/metricsagent/internal/storage"
)

var (
	addedMetrics map[string]metrics.Metric
)

func UpdateMetricHandler(f *os.File) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		mName := chi.URLParam(r, "metricName")
		mType := chi.URLParam(r, "metricType")
		mValue := chi.URLParam(r, "metricValue")

		if strings.Compare(mType, "gauge") != 0 && strings.Compare(mType, "counter") != 0 {
			http.Error(rw, "", http.StatusNotImplemented)
			return
		}

		metricURIValues := make([]string, 0)
		metricURIValues = append(metricURIValues, mType)
		metricURIValues = append(metricURIValues, mName)
		metricURIValues = append(metricURIValues, mValue)

		m, ok := getMetricFromValues(metricURIValues)
		if !ok && m == nil {
			http.Error(rw, "Metric wont't found", http.StatusBadRequest)
			return
		}

		switch m.GetTypeName() {
		case metrics.CounterType:
			value, err := strconv.ParseInt(mValue, 0, 64)
			if err != nil {
				http.Error(rw, "", http.StatusBadRequest)
				return
			}
			m.UpdateValue(value)
		case metrics.GaugeType:
			value, err := strconv.ParseFloat(mValue, 64)
			if err != nil {
				http.Error(rw, "", http.StatusBadRequest)
				return
			}
			m.UpdateValue(value)
		}

		addedMetrics[m.GetName()] = m
		storage.WriteMetricsToFile(f, addedMetrics)

		rw.Header().Add("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusOK)
	}
}

func GetMetricValueFromServer(f *os.File) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "metricType")
		mName := chi.URLParam(r, "metricName")

		metricURIValues := make([]string, 0)
		metricURIValues = append(metricURIValues, mType)
		metricURIValues = append(metricURIValues, mName)

		m, ok := addedMetrics[mName]
		if !ok {
			http.Error(rw, "Metric wasn't found", http.StatusNotFound)
			return
		}

		mValue := m.GetValue()
		switch v := mValue.(type) {
		/* case metrics.GaugeMetric:
			rw.Write([]byte(fmt.Sprintf("%v", v)))
		case metrics.CounterMetric:
			rw.Write([]byte(fmt.Sprintf("%v", v))) */
		default:
			rw.Write([]byte(fmt.Sprintf("%v", v)))
		}
		rw.WriteHeader(http.StatusOK)
	}
}

func getMetricFromValues(sendedValues []string) (metrics.Metric, bool) {
	mType := sendedValues[0]
	metricName := sendedValues[1]

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

	return m, true
}

func SetMetrics(metrics map[string]metrics.Metric) {
	addedMetrics = metrics
}
