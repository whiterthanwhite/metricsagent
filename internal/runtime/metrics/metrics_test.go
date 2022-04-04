package metrics

import (
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
