package metrics

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
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
	Hash  string     `json:"hash,omitempty"`
}

func (m *Metrics) GenerateHash(key string) {
	if key == "" {
		return
	}
	h := hmac.New(sha256.New, []byte(key))
	metricHashString := ""
	switch m.MType {
	case CounterType:
		metricHashString = fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta)
	case GaugeType:
		metricHashString = fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value)
	}
	h.Write([]byte(metricHashString))
	dst := h.Sum(nil)
	m.Hash = string(dst)
}

type Metric interface {
	GetName() string
	GetTypeName() metrictype
	GetValue() interface{}
	UpdateValue(interface{})
	CreateNewMetric() Metrics
}

type GaugeMetric struct {
	Name     string
	TypeName metrictype
	Value    gauge
	mu       sync.RWMutex
}

type CounterMetric struct {
	Name     string
	TypeName metrictype
	Value    counter
	mu       sync.RWMutex
}

func (gm *GaugeMetric) GetName() string {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.Name
}

func (gm *GaugeMetric) GetTypeName() metrictype {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.TypeName
}

func (gm *GaugeMetric) GetValue() interface{} {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return float64(gm.Value)
}

func (gm *GaugeMetric) CreateNewMetric() Metrics {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	newM := Metrics{
		ID:    gm.GetName(),
		MType: gm.GetTypeName(),
	}

	v := gm.GetValue().(float64)
	newM.Value = &v

	return newM
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
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.Name
}

func (cm *CounterMetric) GetTypeName() metrictype {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.TypeName
}

func (cm *CounterMetric) GetValue() interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
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

func (cm *CounterMetric) CreateNewMetric() Metrics {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	newM := Metrics{
		ID:    cm.GetName(),
		MType: cm.GetTypeName(),
	}

	v := cm.GetValue().(int64)
	newM.Delta = &v

	return newM
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
		tempMetric := Metrics{
			ID:    metricDescription.MName,
			MType: metricDescription.MType,
		}
		switch tempMetric.MType {
		case CounterType:
			var value int64 = 0
			tempMetric.Delta = &value
		case GaugeType:
			value := 0.0
			tempMetric.Value = &value
		}
		standardMetrics[metricDescription.MName] = tempMetric
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
