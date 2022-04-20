package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

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
		nM, ok := newMetrics[m.GetName()]
		if !ok {
			tempMetric := metrics.Metrics{
				ID:    m.GetName(),
				MType: m.GetTypeName(),
			}
			switch v := m.GetValue().(type) {
			case int64:
				tempMetric.Delta = &v
			case float64:
				tempMetric.Value = &v
			}
			newMetrics[tempMetric.ID] = tempMetric
		} else {
			switch v := m.GetValue().(type) {
			case int64:
				nM.Delta = &v
			case float64:
				nM.Value = &v
			}
		}

		log.Println(m, nM)
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
		rw.Write(metricsBytes)
	}
}

func UpdateMetricOnServer(serverMetrics map[string]metrics.Metrics) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		log.Println(r.URL)
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(rw, "", http.StatusBadRequest)
			return
		}

		if r.ContentLength == 0 {
			http.Error(rw, "", http.StatusBadRequest)
			return
		}

		var requestMetric metrics.Metrics
		/*
			if err := json.NewDecoder(r.Body).Decode(&requestMetric); err != nil {
				http.Error(rw, fmt.Sprint(err), http.StatusInternalServerError)
				return
			}
			r.Body.Close()
		*/

		requestBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(rw, fmt.Sprint(err), http.StatusBadRequest)
			return
		}
		r.Body.Close()
		log.Println(string(requestBody))

		if err := json.Unmarshal(requestBody, &requestMetric); err != nil {
			http.Error(rw, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
		log.Println(requestMetric)

		m, ok := serverMetrics[requestMetric.ID]
		if !ok {
			serverMetrics[requestMetric.ID] = requestMetric
		} else {
			switch m.MType {
			case metrics.CounterType:
				/*
					mDelta := *m.Delta
					mDelta += *requestMetric.Delta
					m.Delta = &mDelta
				*/
			case metrics.GaugeType:
				m.Value = requestMetric.Value
			}
			serverMetrics[requestMetric.ID] = m
		}

		log.Println("Update OK")
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
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
