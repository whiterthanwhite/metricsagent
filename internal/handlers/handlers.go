package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
	"github.com/whiterthanwhite/metricsagent/internal/storage"
)

func UpdateMetricHandler(f *os.File, addedMetrics map[string]metrics.Metric) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		mName := chi.URLParam(r, "metricName")
		mType := chi.URLParam(r, "metricType")
		mValue := chi.URLParam(r, "metricValue")

		switch mType {
		case "gauge", "counter":
		default:
			http.Error(rw, "", http.StatusNotImplemented)
			return
		}

		metricURIValues := make([]string, 0)
		metricURIValues = append(metricURIValues, mType)
		metricURIValues = append(metricURIValues, mName)
		metricURIValues = append(metricURIValues, mValue)

		m, ok := getMetricFromValues(metricURIValues, addedMetrics)
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

func GetMetricValueFromServer(f *os.File, addedMetrics map[string]metrics.Metric) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		mName := chi.URLParam(r, "metricName")

		m, ok := addedMetrics[mName]
		if !ok {
			http.Error(rw, "Metric wasn't found", http.StatusNotFound)
			return
		}

		rw.WriteHeader(http.StatusOK)
		mValue := m.GetValue()
		switch v := mValue.(type) {
		/* case metrics.GaugeMetric:
			rw.Write([]byte(fmt.Sprintf("%v", v)))
		case metrics.CounterMetric:
			rw.Write([]byte(fmt.Sprintf("%v", v))) */
		default:
			responseWriterWriteCheck(rw, []byte(fmt.Sprintf("%v", v)))
		}
	}
}

func GetAllMetricsFromFile(addedMetrics map[string]metrics.Metric) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		responseWriterWriteCheck(rw, []byte("<html><body>"))
		for _, m := range addedMetrics {
			responseWriterWriteCheck(rw, []byte(fmt.Sprintf(
				"<p>Name: %v; Value: %v</p><br>",
				m.GetName(), m.GetValue())))
		}
		responseWriterWriteCheck(rw, []byte("</body></html>"))
		rw.WriteHeader(http.StatusOK)
	}
}

func getMetricFromValues(sendedValues []string, addedMetrics map[string]metrics.Metric) (metrics.Metric, bool) {
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

func responseWriterWriteCheck(rw http.ResponseWriter, v []byte) {
	_, err := rw.Write(v)
	if err != nil {
		log.Fatal()
	}
}

// new functions
func GetAllMetricsFromServer(serverMetrics []metrics.NewMetric) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(rw, "", http.StatusBadRequest)
		}
		metricsBytes, err := json.Marshal(serverMetrics)
		if err != nil {
			http.Error(rw, "", http.StatusBadRequest)
		}

		_, err = rw.Write(metricsBytes)
		if err != nil {
			http.Error(rw, "", http.StatusInternalServerError)
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.Write(metricsBytes)
	}
}

func GetMetricFromServer(serverMetrics *[]metrics.NewMetric) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		tempServerMetrics := *serverMetrics
		log.Println("GetMetricFromServer")
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(rw, "", http.StatusBadRequest)
			return
		}
		requestBodyBytes, err := getRequestBody(r)
		if err != nil {
			http.Error(rw, "", http.StatusBadRequest)
			return
		}
		if len(requestBodyBytes) == 0 {
			http.Error(rw, "", http.StatusBadRequest)
			return
		}
		requestedMetrics := make([]metrics.NewMetric, 0)
		requestedMetrics = append(requestedMetrics, metrics.NewMetric{})
		/*
			if err := json.Unmarshal(requestBodyBytes, &requestedMetrics); err != nil {
				http.Error(rw, "", http.StatusInternalServerError)
				return
			}
		*/
		if err := json.Unmarshal(requestBodyBytes, &requestedMetrics[0]); err != nil {
			http.Error(rw, "", http.StatusInternalServerError)
			return
		}
		for i := 0; i < len(requestedMetrics); i++ {
			requestedMetric := &requestedMetrics[i]
			for _, serverMetric := range tempServerMetrics {
				log.Println(serverMetric)
				if serverMetric.ID == (*requestedMetric).ID && serverMetric.MType == (*requestedMetric).MType {
					(*requestedMetric).Delta = serverMetric.Delta
					(*requestedMetric).Value = serverMetric.Value
				}
			}
		}
		requestedMetricsBytes, err := json.Marshal(requestedMetrics)
		if err != nil {
			http.Error(rw, "", http.StatusInternalServerError)
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.Write(requestedMetricsBytes)
	}
}

func UpdateMetricOnServer(serverMetrics *[]metrics.NewMetric) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(rw, "", http.StatusBadRequest)
		}
		requestBodyBytes, err := getRequestBody(r)
		if err != nil {
			http.Error(rw, "", http.StatusBadRequest)
		}
		var updateMetrics []metrics.NewMetric
		if err := json.Unmarshal(requestBodyBytes, &updateMetrics); err != nil {
			http.Error(rw, "", http.StatusInternalServerError)
			return
		}
		if len(updateMetrics) == 0 {
			http.Error(rw, "", http.StatusBadRequest)
			return
		}
		tempMetrics := *serverMetrics
		for i := 0; i < len(updateMetrics); i++ {
			updateMetric := updateMetrics[i]
			metricFound := false
			for j := 0; j < len(tempMetrics); j++ {
				serverMetric := &tempMetrics[i]
				if (*serverMetric).ID == updateMetric.ID && (*serverMetric).MType == updateMetric.MType {
					serverMetric.Delta = updateMetric.Delta
					serverMetric.Value = updateMetric.Value
					metricFound = true
				}
			}
			if !metricFound {
				tempMetrics = append(tempMetrics, metrics.NewMetric{
					ID:    updateMetric.ID,
					MType: updateMetric.MType,
					Delta: updateMetric.Delta,
					Value: updateMetric.Value,
				})
			}
		}
		*serverMetrics = tempMetrics
		rw.Header().Set("Content-Type", "application/json")
	}
}

func getRequestBody(r *http.Request) ([]byte, error) {
	requestBody, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, err
	}
	log.Println(string(requestBody))
	return requestBody, nil
}
