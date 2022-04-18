package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
)

func TestGetMetricFromServer(t *testing.T) {
	type send struct {
		contentType string
		testMetrics []metrics.NewMetric
	}
	type want struct {
		contentType   string
		statusCode    int
		metricValues  []float64
		serverMetrics []metrics.NewMetric
	}
	tests := []struct {
		name    string
		request string
		send    send
		want    want
	}{
		{
			name:    "test1",
			request: "/value/",
			send: send{
				contentType: "application/json",
				testMetrics: []metrics.NewMetric{
					{
						ID:    "testGaugeMetric01",
						MType: "gauge",
					},
					{
						ID:    "testGaugeMetric02",
						MType: "gauge",
					},
				},
			},
			want: want{
				contentType:  "application/json",
				statusCode:   200,
				metricValues: []float64{150.01, 1003.405},
				serverMetrics: []metrics.NewMetric{
					{
						ID:    "testGaugeMetric01",
						MType: "gauge",
					},
					{
						ID:    "testGaugeMetric02",
						MType: "gauge",
					},
				},
			},
		},
		{
			name:    "test 2",
			request: "/value/",
			send: send{
				contentType: "application/json",
				testMetrics: []metrics.NewMetric{
					{
						ID:    "testGaugeMetric01",
						MType: "gauge",
					},
				},
			},
			want: want{
				contentType:  "application/json",
				statusCode:   200,
				metricValues: []float64{},
				serverMetrics: []metrics.NewMetric{
					{
						ID:    "testGaugeMetric01",
						MType: "gauge",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, metricValue := range tt.want.metricValues {
				tt.want.serverMetrics[i].Value = &metricValue
			}
			jsonServerMetric, err := json.Marshal(tt.want.serverMetrics)
			if err != nil {
				panic(err)
			}

			testMetrics := tt.send.testMetrics
			requestBodyBytes, err := json.Marshal(testMetrics)
			if err != nil {
				panic(err)
			}
			requestBodyBuffer := bytes.NewBuffer(requestBodyBytes)
			request := httptest.NewRequest(http.MethodPost, tt.request, requestBodyBuffer)
			request.Header.Set("Content-Type", tt.send.contentType)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(GetMetricFromServer(tt.want.serverMetrics))
			h.ServeHTTP(w, request)
			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			userResult, err := ioutil.ReadAll(result.Body)
			defer result.Body.Close()
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, string(jsonServerMetric), string(userResult))
		})
	}
}

func TestUpdateMetricOnServer(t *testing.T) {
	type (
		send struct {
			contentType   string
			updateValues  []float64
			updateMetrics []metrics.NewMetric
		}
		want struct {
			status      int
			contentType string
		}
	)
	tests := []struct {
		name          string
		target        string
		serverMetrics []metrics.NewMetric
		send          send
		want          want
	}{
		{
			name:   "test1",
			target: "/update/",
			send: send{
				contentType:  "application/json",
				updateValues: []float64{150.01},
				updateMetrics: []metrics.NewMetric{
					{
						ID:    "testGaugeMetric01",
						MType: "gauge",
					},
				},
			},
			want: want{
				status:      200,
				contentType: "application/json",
			},
		},
		{
			name:   "test2",
			target: "/update/",
			send: send{
				contentType:  "application/json",
				updateValues: []float64{949839.1573033818},
				updateMetrics: []metrics.NewMetric{
					{
						ID:    "GetSet244",
						MType: "gauge",
					},
				},
			},
			want: want{
				status:      200,
				contentType: "application/json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateMetrics := tt.send.updateMetrics
			for i, updateValue := range tt.send.updateValues {
				updateMetrics[i].Value = &updateValue
			}
			updateMetricBytes, err := json.Marshal(updateMetrics)
			if err != nil {
				panic(err)
			}
			updateMetricsBuffer := bytes.NewBuffer(updateMetricBytes)
			request := httptest.NewRequest(http.MethodPost, tt.target, updateMetricsBuffer)
			request.Header.Set("Content-Type", tt.send.contentType)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(UpdateMetricOnServer(&tt.serverMetrics))
			h.ServeHTTP(w, request)
			result := w.Result()
			result.Body.Close()

			assert.Equal(t, tt.want.status, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			serverMetricsBytes, err := json.Marshal(tt.serverMetrics)
			if err != nil {
				panic(err)
			}
			assert.Equal(t, string(updateMetricBytes), string(serverMetricsBytes))
		})
	}
}
