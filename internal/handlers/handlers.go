package handlers

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/whiterthanwhite/metricsagent/internal/metricdb"
	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
	"github.com/whiterthanwhite/metricsagent/internal/settings"
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

func GetAllMetricsFromFile(addedMetrics map[string]metrics.Metric, newServerMetrics map[string]metrics.Metrics) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/html")
		if r.Header.Get("Accept-Encoding") == "gzip" {
			rw.Header().Set("Content-Encoding", "gzip")
			var returnBuffer bytes.Buffer
			gzipW := gzip.NewWriter(&returnBuffer)
			if _, err := gzipW.Write([]byte("<html><body>")); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			for _, m := range newServerMetrics {
				htmlString := ""
				switch m.MType {
				case metrics.CounterType:
					htmlString = fmt.Sprintf("<p>Name: %v; Value: %v</p><br>", m.ID, *m.Delta)
				case metrics.GaugeType:
					htmlString = fmt.Sprintf("<p>Name: %v; Value: %v</p><br>", m.ID, *m.Value)
				}
				if _, err := gzipW.Write([]byte(htmlString)); err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if _, err := gzipW.Write([]byte("</body></html>")); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err := rw.Write(returnBuffer.Bytes()); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			responseWriterWriteCheck(rw, []byte("<html><body>"))
			for _, m := range addedMetrics {
				responseWriterWriteCheck(rw, []byte(fmt.Sprintf(
					"<p>Name: %v; Value: %v</p><br>",
					m.GetName(), m.GetValue())))
			}
			responseWriterWriteCheck(rw, []byte("</body></html>"))
		}
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

		writeResponseBody(metricsBytes, rw)
	}
}

func UpdateMetricOnServer(serverMetrics map[string]metrics.Metrics, serverSettings settings.SysSettings) http.HandlerFunc {
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
		if err := getMetricFromRequestBody(&requestMetric, r); err != nil {
			http.Error(rw, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}

		// Check hash key >>
		if serverSettings.Key != "" {
			requestHash, err := hex.DecodeString(requestMetric.Hash)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			s := requestMetric.GenerateHash(serverSettings.Key)
			checkHash, err := hex.DecodeString(s)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			if !hmac.Equal(requestHash, checkHash) {
				http.Error(rw, "", http.StatusBadRequest)
				return
			}
		}
		// Check hash key <<

		m, ok := serverMetrics[requestMetric.ID]
		if !ok {
			serverMetrics[requestMetric.ID] = requestMetric
			m = requestMetric
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

		writeResponseBody([]byte("{}"), rw)
	}
}

func GetMetricFromServer(serverMetrics map[string]metrics.Metrics, serverSettings settings.SysSettings) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var requestMetric metrics.Metrics
		if err := getMetricFromRequestBody(&requestMetric, r); err != nil {
			http.Error(rw, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}

		m, ok := serverMetrics[requestMetric.ID]
		if !ok {
			http.Error(rw, "metric is not found", http.StatusNotFound)
			return
		}

		m.Hash = m.GenerateHash(serverSettings.Key)
		returnMetric, err := json.Marshal(m)
		if err != nil {
			http.Error(rw, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Accept-Encoding") == "gzip" {
			rw.Header().Set("Content-Encoding", "gzip")
			var encodedBuffer bytes.Buffer
			gzipW := gzip.NewWriter(&encodedBuffer)
			if _, err := gzipW.Write(returnMetric); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			if err := gzipW.Close(); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			returnMetric = encodedBuffer.Bytes()
		}

		writeResponseBody(returnMetric, rw)
	}
}

func writeResponseBody(v []byte, rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "application/json")
	if _, err := rw.Write(v); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func getMetricFromRequestBody(m *metrics.Metrics, r *http.Request) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(m); err != nil {
		return err
	}
	return nil
}

func CheckDatabaseConn(conn metricdb.Metricdb) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if err := conn.Ping(); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(http.StatusOK)
	}
}
