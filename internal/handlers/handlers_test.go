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
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, string(jsonServerMetric), string(userResult))

		})
	}
}
