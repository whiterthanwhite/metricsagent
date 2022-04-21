package metrics

import (
	"log"
	"strconv"
	"strings"
	"sync"
)

type (
	metrictype string
	gauge      float64
	counter    int64
)

const (
	GaugeType   metrictype = "gauge"
	CounterType metrictype = "counter"
)

type Metrics struct {
	ID    string     `json:"id"`
	MType metrictype `json:"type"`
	Delta *int64     `json:"delta,omitempty"`
	Value *float64   `json:"value,omitempty"`
}

type Metric interface {
	GetName() string
	GetTypeName() metrictype
	GetValue() interface{}
	UpdateValue(interface{})
}

type GaugeMetric struct {
	Name     string
	TypeName metrictype
	Value    gauge
	mu       sync.Mutex
}

type CounterMetric struct {
	Name     string
	TypeName metrictype
	Value    counter
	mu       sync.Mutex
}

func (gm *GaugeMetric) GetName() string {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	return gm.Name
}

func (gm *GaugeMetric) GetTypeName() metrictype {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	return gm.TypeName
}

func (gm *GaugeMetric) GetValue() interface{} {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	return float64(gm.Value)
}

func (gm *GaugeMetric) UpdateValue(v interface{}) {
	newValue, ok := v.(float64)
	if ok {
		gm.mu.Lock()
		gm.Value = gauge(newValue)
		gm.mu.Unlock()
	}
}

func (cm *CounterMetric) GetName() string {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.Name
}

func (cm *CounterMetric) GetTypeName() metrictype {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.TypeName
}

func (cm *CounterMetric) GetValue() interface{} {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return int64(cm.Value)
}

func (cm *CounterMetric) UpdateValue(v interface{}) {
	newValue, ok := v.(int64)
	if ok {
		cm.mu.Lock()
		cm.Value += counter(newValue)
		cm.mu.Unlock()
	}
}

func GetAllMetrics() map[string]Metric {
	metrics := make(map[string]Metric)
	metricDescriptions := GetStandardMetrics()
	for _, mDescription := range metricDescriptions {
		metrics[mDescription.MName] = createMetric(mDescription.MName, mDescription.MType)
	}

	return metrics
}

func GetAllNewMetrics() map[string]Metrics {
	metricsDescription := GetStandardMetrics()
	standardMetrics := make(map[string]Metrics)
	for _, metricDescription := range metricsDescription {
		standardMetrics[metricDescription.MName] = Metrics{
			ID:    metricDescription.MName,
			MType: metricDescription.MType,
		}
	}
	return standardMetrics
}

func GetMetric(name string, mType string) Metric {
	mt := metrictype(mType)
	return createMetric(name, mt)
}

func createMetric(name string, kind metrictype) Metric {
	switch kind {
	case GaugeType:
		return &GaugeMetric{
			Name:     name,
			TypeName: kind,
		}
	case CounterType:
		return &CounterMetric{
			Name:     name,
			TypeName: kind,
		}
	}
	return nil
}

func IsMetricTypeExist(mType string) bool {
	switch metrictype(mType) {
	case GaugeType, CounterType:
		return true
	}
	return false
}

func ParseCSVString(csvStr string) Metric {
	metricValues := strings.Split(csvStr, ";")
	mt := metrictype(metricValues[0])
	switch mt {
	case GaugeType:
		tempFloat, err := strconv.ParseFloat(metricValues[2], 64)
		if err != nil {
			log.Fatal(err)
		}
		return &GaugeMetric{
			Name:     metricValues[1],
			TypeName: mt,
			Value:    gauge(tempFloat),
		}
	case CounterType:
		tempInt, err := strconv.ParseInt(metricValues[2], 0, 64)
		if err != nil {
			log.Fatal(err)
		}
		return &CounterMetric{
			Name:     metricValues[1],
			TypeName: mt,
			Value:    counter(tempInt),
		}
	}

	return nil
}
