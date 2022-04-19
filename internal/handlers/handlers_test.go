package handlers

import (
	// "io/ioutil"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	// "github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/require"

	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
)

func TestUpdateMetricOnServer(t *testing.T) {
	serverMetrics := make([]metrics.NewMetric, 0)
	type send struct {
		m      metrics.NewMetric
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
				m:      metrics.NewMetric{ID: "Metric 1", MType: metrics.CounterType},
				mDelta: 1,
				mValue: 0.0,
			},
		},
		{
			name: "test 2",
			send: send{
				m:      metrics.NewMetric{ID: "Metric 2", MType: metrics.CounterType},
				mDelta: 2,
				mValue: 0.0,
			},
		},
		{
			name: "test 3",
			send: send{
				m:      metrics.NewMetric{ID: "Metric 3", MType: metrics.CounterType},
				mDelta: 2,
				mValue: 0.0,
			},
		},
		{
			name: "test 4",
			send: send{
				m:      metrics.NewMetric{ID: "Metric 4", MType: metrics.GaugeType},
				mDelta: 0,
				mValue: 0.01,
			},
		},
		{
			name: "test 5",
			send: send{
				m:      metrics.NewMetric{ID: "Metric 4", MType: metrics.GaugeType},
				mDelta: 0,
				mValue: 0.02,
			},
		},
		{
			name: "test 6",
			send: send{
				m:      metrics.NewMetric{ID: "Metric 5", MType: metrics.GaugeType},
				mDelta: 0,
				mValue: 0.03,
			},
		},
		{
			name: "test 7",
			send: send{
				m:      metrics.NewMetric{ID: "Metric 6", MType: metrics.CounterType},
				mDelta: 0,
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

			request := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(rM))
			request.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h := http.HandlerFunc(UpdateMetricOnServer(&serverMetrics))
			h.ServeHTTP(w, request)
			result := w.Result()
			result.Body.Close()
		})
	}
}
