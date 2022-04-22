package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	// "github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"

	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
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
	serverMetrics := make(map[string]metrics.Metrics)
	type send struct {
		m      metrics.Metrics
		mDelta int64
		mValue float64
	}

	tests := []struct {
		name string
		send send
	}{
		{
			name: "test 1",
			send: send{
				m:      metrics.Metrics{ID: "Metric 1", MType: metrics.CounterType},
				mDelta: 1,
				mValue: 0.0,
			},
		},
		{
			name: "test 2",
			send: send{
				m:      metrics.Metrics{ID: "Metric 2", MType: metrics.CounterType},
				mDelta: 2,
				mValue: 0.0,
			},
		},
		{
			name: "test 3",
			send: send{
				m:      metrics.Metrics{ID: "Metric 3", MType: metrics.CounterType},
				mDelta: 2,
				mValue: 0.0,
			},
		},
		{
			name: "test 4",
			send: send{
				m:      metrics.Metrics{ID: "Metric 4", MType: metrics.GaugeType},
				mDelta: 0,
				mValue: 0.01,
			},
		},
		{
			name: "test 5",
			send: send{
				m:      metrics.Metrics{ID: "Metric 4", MType: metrics.GaugeType},
				mDelta: 0,
				mValue: 0.02,
			},
		},
		{
			name: "test 6",
			send: send{
				m:      metrics.Metrics{ID: "Metric 5", MType: metrics.GaugeType},
				mDelta: 0,
				mValue: 0.03,
			},
		},
		{
			name: "test 7",
			send: send{
				m:      metrics.Metrics{ID: "Metric 6", MType: metrics.CounterType},
				mDelta: 1,
				mValue: 0.0,
			},
		},
		{
			name: "test 8",
			send: send{
				m:      metrics.Metrics{ID: "Metric 10", MType: metrics.CounterType},
				mDelta: 1,
				mValue: 0.0,
			},
		},
		{
			name: "test 9",
			send: send{
				m:      metrics.Metrics{ID: "Metric 1", MType: metrics.CounterType},
				mDelta: 1,
				mValue: 0.0,
			},
		},
		{
			name: "test 10",
			send: send{
				m:      metrics.Metrics{ID: "Metric 1", MType: metrics.CounterType},
				mDelta: 1,
				mValue: 0.0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.send.mDelta != 0 {
				tt.send.m.Delta = &tt.send.mDelta
			}
			if tt.send.mValue != 0 {
				tt.send.m.Value = &tt.send.mValue
			}

			rM, err := json.Marshal(tt.send.m)
			if err != nil {
				panic(err)
			}

			//rM = append(rM, 40)
			log.Println(string(rM))

			request := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(rM))
			request.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h := http.HandlerFunc(UpdateMetricOnServer(serverMetrics))
			h.ServeHTTP(w, request)
			result := w.Result()
			result.Body.Close()

			m, ok := serverMetrics[tt.send.m.ID]
			if !ok {
				panic(ok)
			}
			switch m.MType {
			case metrics.CounterType:
				log.Println(m.ID, m.MType, *m.Delta)
			case metrics.GaugeType:
				log.Println(m.ID, m.MType, *m.Value)
			}
		})
	}
}

func TestGetMetricFromServer(t *testing.T) {
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
			h := http.HandlerFunc(GetMetricFromServer(serverMetrics))
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
