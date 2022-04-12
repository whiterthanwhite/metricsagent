package handlers

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
	"github.com/whiterthanwhite/metricsagent/internal/storage"
)

func TestGetMetricValueFromServer(t *testing.T) {
	_, cancel := context.WithCancel(context.Background())
	time.AfterFunc(1*time.Second, cancel)

	metricFile := storage.OpenMetricFileCSV()
	addedMetrics := storage.GetMetricsFromFile(metricFile)
	if len(addedMetrics) == 0 {
		addedMetrics = metrics.GetAllMetrics()
	}
	defer metricFile.Close()

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Get("/", GetAllMetricsFromFile())
		r.Route("/update", func(r chi.Router) {
			r.Post("/{metricType}/{metricName}/{metricValue}",
				UpdateMetricHandler(metricFile, addedMetrics))
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/{metricType}/{metricName}",
				GetMetricValueFromServer(metricFile, addedMetrics))
		})
	})
	ts := httptest.NewServer(r)
	defer ts.Close()

	var resp *http.Response
	var body string

	resp, _ = getMetricValueFromServer(t, ts, http.MethodPost, "/update/counter/testCounter/100")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "")
	resp.Body.Close()

	resp, _ = getMetricValueFromServer(t, ts, http.MethodPost, "/update/counter/testCounter/none")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()

	resp, _ = getMetricValueFromServer(t, ts, http.MethodPost, "/update/gauge/testGauge/100")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp, _ = getMetricValueFromServer(t, ts, http.MethodPost, "/update/counter/testSetGet33/527")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp, _ = getMetricValueFromServer(t, ts, http.MethodPost, "/update/counter/testSetGet33/455")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp, _ = getMetricValueFromServer(t, ts, http.MethodPost, "/update/counter/testSetGet33/187")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp, _ = getMetricValueFromServer(t, ts, http.MethodPost, "/update/gauge/testSetGet134/65637.019")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp, body = getMetricValueFromServer(t, ts, http.MethodGet, "/value/gauge/testSetGet134")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "65637.019", body)
	resp.Body.Close()

	resp, _ = getMetricValueFromServer(t, ts, http.MethodPost, "/update/counter/testSetGet33/527")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp, body = getMetricValueFromServer(t, ts, http.MethodGet, "/value/counter/testSetGet33")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "1696", body)
	resp.Body.Close()

	resp, _ = getMetricValueFromServer(t, ts, http.MethodPost, "/update/counter/testSetGet33/982")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp, body = getMetricValueFromServer(t, ts, http.MethodGet, "/value/counter/testSetGet33")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "2678", body)
	resp.Body.Close()

	resp, _ = getMetricValueFromServer(t, ts, http.MethodPost, "/update/counter/testSetGet33/1169")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp, body = getMetricValueFromServer(t, ts, http.MethodGet, "/value/counter/testSetGet33")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "3847", body)
	resp.Body.Close()
}

func getMetricValueFromServer(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}
