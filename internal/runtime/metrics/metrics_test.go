package metrics

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMetricTypeExist(t *testing.T) {
	tests := []struct {
		name     string
		mType    string
		expected bool
	}{
		{
			name:     "test 1",
			mType:    "gauge",
			expected: true,
		},
		{
			name:     "test 2",
			mType:    "counter",
			expected: true,
		},
		{
			name:     "test 2",
			mType:    "unknown",
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsMetricTypeExist(tt.mType))
		})
	}
}

func TestGetMetric(t *testing.T) {
	tests := []struct {
		name           string
		expectedMetric Metric
		mName          string
		mType          string
	}{
		{
			name: "test 1",
			expectedMetric: &GaugeMetric{
				Name:     "Alloc",
				TypeName: GaugeType,
			},
			mName: "Alloc",
			mType: "gauge",
		},
		{
			name: "test 2",
			expectedMetric: &CounterMetric{
				Name:     "PollCount",
				TypeName: CounterType,
			},
			mName: "PollCount",
			mType: "counter",
		},
		{
			name: "test 3",
			expectedMetric: &CounterMetric{
				Name:     "MyCounter",
				TypeName: CounterType,
			},
			mName: "MyCounter",
			mType: "counter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedMetric, GetMetric(tt.mName, tt.mType))
		})
	}
}

func TestGenerateHash(t *testing.T) {
	type want struct {
		key string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "test 1",
			want: want{
				key: "key1",
			},
		},
		{
			name: "test 2",
			want: want{
				key: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			value := 134.12345
			newMetric := Metrics{
				ID:    "Alloc",
				MType: GaugeType,
				Value: &value,
			}

			if tt.want.key != "" {
				h := hmac.New(sha256.New, []byte(tt.want.key))
				metricHashString := ""
				switch newMetric.MType {
				case CounterType:
					metricHashString = fmt.Sprintf("%s:counter:%d", newMetric.ID, *newMetric.Delta)
				case GaugeType:
					metricHashString = fmt.Sprintf("%s:gauge:%f", newMetric.ID, *newMetric.Value)
				}
				h.Write([]byte(metricHashString))
				dst := h.Sum(nil)

				newMetric.Hash = newMetric.GenerateHash(tt.want.key)

				dst2, err := hex.DecodeString(newMetric.Hash)
				assert.Nil(t, err)
				assert.True(t, hmac.Equal(dst, dst2))
			} else {
				newMetric.Hash = newMetric.GenerateHash(tt.want.key)
				assert.Equal(t, tt.want.key, newMetric.Hash)
			}
		})
	}
}
