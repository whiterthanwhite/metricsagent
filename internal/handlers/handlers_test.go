package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/whiterthanwhite/metricsagent/internal/metricdb"
	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
	"github.com/whiterthanwhite/metricsagent/internal/settings"
)

func TestUpdateMetricHandler(t *testing.T) {
	newMetrics := make(map[string]metrics.Metrics)
	oldMetrics := make(map[string]metrics.Metric)

	tests := []struct {
		name   string
		target string
	}{
		{
			name:   "test 1",
			target: "/gauge/Sys/25",
		},
		{
			name:   "test 2",
			target: "/counter/PollCount/25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.target, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(UpdateMetricHandler(oldMetrics, newMetrics))
			h.ServeHTTP(w, request)
			result := w.Result()
			result.Body.Close()

			for _, oldMetric := range oldMetrics {
				log.Println(oldMetric, oldMetric.GetValue())
			}
			for _, newMetric := range newMetrics {
				switch newMetric.MType {
				case metrics.CounterType:
					log.Println(newMetric, *newMetric.Delta)
				case metrics.GaugeType:
					log.Println(newMetric, *newMetric.Value)
				}
			}
		})
	}
}

func TestUpdateMetricOnServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serverSettings := settings.GetSysSettings()
	conn := metricdb.CreateConnnection(ctx, "")
	if conn.IsConnClose() {
		return
	}
	defer func() {
		if err := conn.CloseConnection(ctx); err != nil {
			log.Println(err.Error())
		}
	}()

	serverMetrics := make(map[string]metrics.Metrics)

	type want struct {
		statusCode int
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "test 1",
			want: want{
				statusCode: 200,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v int64 = 4
			m := metrics.Metrics{
				ID:    "Alloc",
				MType: metrics.CounterType,
				Delta: &v,
			}

			mb, err := json.Marshal(m)
			assert.Nil(t, err)

			buff := bytes.NewBuffer(mb)
			request := httptest.NewRequest(http.MethodPost, "/update", buff)
			request.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h := http.HandlerFunc(UpdateMetricOnServer(serverMetrics, serverSettings, conn))
			h.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			body, err := io.ReadAll(result.Body)
			assert.Nil(t, err)
			log.Println(string(body))

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
		})
	}
}

func TestUpdateMetricsOnServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serverSettings := settings.GetSysSettings()
	conn := metricdb.CreateConnnection(ctx, serverSettings.MetricDBAdress)
	if conn.IsConnClose() {
		return
	}
	defer func() {
		if err := conn.CloseConnection(ctx); err != nil {
			log.Println(err.Error())
		}
	}()

	serverMetrics := make(map[string]metrics.Metrics)

	type want struct {
		statusCode int
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "test 1",
			want: want{
				statusCode: 200,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v int64 = 4
			ms := []metrics.Metrics{
				{
					ID:    "Metric 1",
					MType: metrics.CounterType,
					Delta: &v,
				},
				{
					ID:    "Metric 2",
					MType: metrics.CounterType,
					Delta: &v,
				},
				{
					ID:    "Metric 3",
					MType: metrics.CounterType,
					Delta: &v,
				},
			}

			mb, err := json.Marshal(ms)
			assert.Nil(t, err)

			buff := bytes.NewBuffer(mb)
			request := httptest.NewRequest(http.MethodPost, "/update", buff)
			request.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			h := http.HandlerFunc(UpdateMetricsOnServer(serverMetrics, serverSettings, conn))
			h.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			responseBody, err := io.ReadAll(result.Body)
			assert.Nil(t, err)
			log.Println("Response body: ", string(responseBody))

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
		})
	}
}

func TestGetMetricFromServer(t *testing.T) {
	serverSettings := settings.GetSysSettings()
	serverMetricDeltas := []int64{
		0,
	}
	serverMetrics := make(map[string]metrics.Metrics)
	serverMetrics["Metric 1"] = metrics.Metrics{
		ID:    "Metric 1",
		MType: metrics.CounterType,
		Delta: &serverMetricDeltas[0],
	}

	type (
		send struct {
			m metrics.Metrics
		}
		want struct {
			m      metrics.Metrics
			mDelta int64
			mValue float64
		}
	)

	tests := []struct {
		name string
		send send
		want want
	}{
		{
			name: "test 1",
			send: send{
				m: metrics.Metrics{ID: "Metric 1", MType: metrics.CounterType},
			},
			want: want{
				m:      metrics.Metrics{ID: "Metric 1", MType: metrics.CounterType},
				mDelta: 0,
				mValue: 0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rM, err := json.Marshal(tt.send.m)
			if err != nil {
				panic(err)
			}

			request := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(rM))
			request.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h := http.HandlerFunc(GetMetricFromServer(serverMetrics, serverSettings))
			h.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			rBody, err := ioutil.ReadAll(result.Body)
			if err != nil {
				panic(err)
			}

			switch tt.want.m.MType {
			case metrics.CounterType:
				tt.want.m.Delta = &tt.want.mDelta
			case metrics.GaugeType:
				tt.want.m.Value = &tt.want.mValue
			}

			expectedMetricJSON, err := json.Marshal(tt.want.m)
			if err != nil {
				panic(err)
			}

			assert.Equal(t, string(expectedMetricJSON), string(rBody))
		})
	}
}

func TestCheckDatabaseConn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conn := metricdb.CreateConnnection(ctx, "")
	if conn.IsConnClose() {
		return
	}
	defer func() {
		if err := conn.CloseConnection(ctx); err != nil {
			log.Println(err.Error())
		}
	}()

	type want struct {
		code int
	}

	tests := []struct {
		name string
		want want
	}{
		{
			name: "test 1",
			want: want{
				code: 200,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(CheckDatabaseConn(conn))
			h.ServeHTTP(w, r)
			result := w.Result()
			result.Body.Close()

			assert.Equal(t, tt.want.code, result.StatusCode)
		})
	}
}
