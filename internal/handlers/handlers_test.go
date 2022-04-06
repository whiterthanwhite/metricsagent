package handlers

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/whiterthanwhite/metricsagent/internal/storage"
)

func TestGetMetricValueFromServer(t *testing.T) {
	var f *os.File = storage.OpenMetricFileCSV()
	f.Close()

	r := chi.NewRouter()
	ts := httptest.NewServer(r)
	ts.URL = "http://127.0.0.1:8080"

	defer ts.Close()

	resp, body := testGetMetricValueFromServer(t, ts, http.MethodPost, "/update/unknown/testCounter/100")
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
	// assert.Equal(t, "", body)

	resp, body = testGetMetricValueFromServer(t, ts, http.MethodGet, "/value/counter/testSetGet33")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	// assert.Equal(t, "", body)

	resp, _ = testGetMetricValueFromServer(t, ts, http.MethodGet, "/value/counter/testUnknown15")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	// assert.Equal(t, "404 page not found", body)

	resp, _ = testGetMetricValueFromServer(t, ts, http.MethodPost, "/update/gauge/testSetGet199/437714.187")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// assert.Equal(t, "404 page not found", body)

	resp, body = testGetMetricValueFromServer(t, ts, http.MethodGet, "/value/gauge/testSetGet199")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "437714.187", body)

	os.Remove(f.Name())
}

func testGetMetricValueFromServer(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}
