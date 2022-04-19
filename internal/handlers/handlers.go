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
		log.Println("GetMetricFromServer")
		tempServerMetrics := *serverMetrics
		// Debug >>
		var pollCountVal int64 = 0
		tempServerMetrics = append(tempServerMetrics, metrics.NewMetric{
			ID:    "PollCount",
			MType: metrics.CounterType,
			Delta: &pollCountVal,
		})
		// Debug <<
		log.Println(tempServerMetrics)

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(rw, "", http.StatusBadRequest)
			return
		}

		requestBodyBytes, err := getRequestBody(r)
		if err != nil {
			http.Error(rw, "", http.StatusBadRequest)
			return
		}
		log.Println(string(requestBodyBytes))

		if len(requestBodyBytes) == 0 {
			http.Error(rw, "", http.StatusBadRequest)
			return
		}

		var rM metrics.NewMetric
		if err := json.Unmarshal(requestBodyBytes, &rM); err != nil {
			http.Error(rw, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		log.Println(rM)

		for _, tempServerMetric := range tempServerMetrics {
			if rM.ID == tempServerMetric.ID && rM.MType == tempServerMetric.MType {
				rM.Delta = tempServerMetric.Delta
				rM.Value = tempServerMetric.Value
			}
		}
		log.Println(rM)

		bRM, err := json.Marshal(rM)
		if err != nil {
			http.Error(rw, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		log.Println(string(bRM))

		rw.Header().Set("Content-Type", "application/json")
		rw.Write(bRM)
	}
}

func UpdateMetricOnServer(serverMetrics *[]metrics.NewMetric) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		log.Println("UpdateMetricOnServer")
		tempServerMetrics := *serverMetrics
		log.Println(tempServerMetrics)

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(rw, "", http.StatusBadRequest)
		}

		requestBodyBytes, err := getRequestBody(r)
		if err != nil {
			http.Error(rw, "", http.StatusBadRequest)
		}
		log.Println(string(requestBodyBytes))

		if len(requestBodyBytes) == 0 {
			http.Error(rw, "", http.StatusBadRequest)
			return
		}

		var updateMetric metrics.NewMetric
		if err := json.Unmarshal(requestBodyBytes, &updateMetric); err != nil {
			http.Error(rw, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		log.Println(updateMetric)

		mFound := false
		for _, tempServerMetric := range tempServerMetrics {
			if updateMetric.ID == tempServerMetric.ID && updateMetric.MType == tempServerMetric.MType {
				tempServerMetric.Delta = updateMetric.Delta
				tempServerMetric.Value = updateMetric.Value
				mFound = true
			}
		}
		if !mFound {
			tempServerMetrics = append(tempServerMetrics, updateMetric)
		}

		*serverMetrics = tempServerMetrics
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
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
