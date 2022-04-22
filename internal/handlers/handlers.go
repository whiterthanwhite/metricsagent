package handlers

import (
	"encoding/json"
	"fmt"

	//"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
)

func UpdateMetricHandler(addedMetrics map[string]metrics.Metric, newMetrics map[string]metrics.Metrics) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		mName := chi.URLParam(r, "metricName")
		mType := chi.URLParam(r, "metricType")
		mValue := chi.URLParam(r, "metricValue")

		params := strings.Split(r.URL.String(), "/")
		if mType == "" {
			mType = params[1]
		}
		if mName == "" {
			mName = params[2]
		}
		if mValue == "" {
			mValue = params[3]
		}

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
				http.Error(rw, fmt.Sprint(err), http.StatusBadRequest)
				return
			}
			m.UpdateValue(value)
		case metrics.GaugeType:
			value, err := strconv.ParseFloat(mValue, 64)
			if err != nil {
				http.Error(rw, fmt.Sprint(err), http.StatusBadRequest)
				return
			}
			m.UpdateValue(value)
		}
		addedMetrics[m.GetName()] = m

		nM, ok := newMetrics[m.GetName()]
		if !ok {
			nM = metrics.Metrics{
				ID:    m.GetName(),
				MType: m.GetTypeName(),
			}
		}
		switch v := m.GetValue().(type) {
		case int64:
			nM.Delta = &v
		case float64:
			nM.Value = &v
		}
		newMetrics[nM.ID] = nM

		rw.Header().Add("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusOK)
	}
}

func GetMetricValueFromServer(addedMetrics map[string]metrics.Metric) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		mName := chi.URLParam(r, "metricName")

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
func GetAllMetricsFromServer(serverMetrics []metrics.Metrics) http.HandlerFunc {
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
		if _, err := rw.Write(metricsBytes); err != nil {
			log.Fatal(err)
		}
	}
}

func UpdateMetricOnServer(serverMetrics map[string]metrics.Metrics) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(rw, "", http.StatusBadRequest)
			return
		}

		if r.ContentLength == 0 {
			http.Error(rw, "", http.StatusBadRequest)
			return
		}

		var requestMetric metrics.Metrics
		if err := json.NewDecoder(r.Body).Decode(&requestMetric); err != nil {
			http.Error(rw, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		r.Body.Close()

		m, ok := serverMetrics[requestMetric.ID]
		if !ok {
			serverMetrics[requestMetric.ID] = requestMetric
		} else {
			switch m.MType {
			case metrics.CounterType:
				var tempDelta int64 = 0
				if m.Delta != nil {
					tempDelta = *m.Delta
				}
				tempDelta += *requestMetric.Delta
				m.Delta = &tempDelta
			case metrics.GaugeType:
				m.Value = requestMetric.Value
			}
			serverMetrics[requestMetric.ID] = m
		}
		log.Println(requestMetric, m)

		rw.Header().Set("Content-Type", "application/json")
		_, err := rw.Write([]byte(`{}`))
		if err != nil {
			http.Error(rw, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
	}
}

func GetMetricFromServer(serverMetrics map[string]metrics.Metrics) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var requestMetric metrics.Metrics
		if err := json.NewDecoder(r.Body).Decode(&requestMetric); err != nil {
			http.Error(rw, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		r.Body.Close()

		m, ok := serverMetrics[requestMetric.ID]
		log.Println(requestMetric, m, ok)
		if !ok {
			http.Error(rw, "metric is not found", http.StatusNotFound)
			return
		}

		returnMetric, err := json.Marshal(m)
		if err != nil {
			http.Error(rw, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		_, err = rw.Write(returnMetric)
		if err != nil {
			http.Error(rw, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
	}
}
