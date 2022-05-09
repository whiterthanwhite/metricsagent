package metrics

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

func (m Metrics) GenerateHash(key string) string {
	if key == "" {
		return ""
	}
	h := hmac.New(sha256.New, []byte(key))
	stringToHash := ""
	switch m.MType {
	case CounterType:
		stringToHash = fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta)
	case GaugeType:
		stringToHash = fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value)
	}
	h.Write([]byte(stringToHash))
	dst := h.Sum(nil)

	return hex.EncodeToString(dst)
}

type Metric interface {
	GetName() string
	GetTypeName() metrictype
	GetValue() interface{}
	UpdateValue(interface{})
	CreateNewMetric() Metrics
}

type GaugeMetric struct {
	mtx      sync.RWMutex
	Name     string
	TypeName metrictype
	Value    gauge
}

type CounterMetric struct {
	mtx      sync.RWMutex
	Name     string
	TypeName metrictype
	Value    counter
}

func (gm *GaugeMetric) GetName() string {
	gm.mtx.RLock()
	defer gm.mtx.RUnlock()

	return gm.Name
}

func (gm *GaugeMetric) GetTypeName() metrictype {
	gm.mtx.RLock()
	defer gm.mtx.RUnlock()

	return gm.TypeName
}

func (gm *GaugeMetric) GetValue() interface{} {
	gm.mtx.RLock()
	defer gm.mtx.RUnlock()

	return float64(gm.Value)
}

func (gm *GaugeMetric) UpdateValue(v interface{}) {
	gm.mtx.Lock()
	defer gm.mtx.Unlock()

	if newValue, ok := v.(float64); ok {
		gm.Value = gauge(newValue)
	}
}

func (gm *GaugeMetric) CreateNewMetric() Metrics {
	newM := Metrics{
		ID:    gm.GetName(),
		MType: gm.GetTypeName(),
	}

	v := gm.GetValue().(float64)
	newM.Value = &v

	return newM
}

func (cm *CounterMetric) GetName() string {
	cm.mtx.RLock()
	defer cm.mtx.RUnlock()

	return cm.Name
}

func (cm *CounterMetric) GetTypeName() metrictype {
	cm.mtx.RLock()
	defer cm.mtx.RUnlock()

	return cm.TypeName
}

func (cm *CounterMetric) GetValue() interface{} {
	cm.mtx.RLock()
	defer cm.mtx.RUnlock()

	return int64(cm.Value)
}

func (cm *CounterMetric) UpdateValue(v interface{}) {
	cm.mtx.Lock()
	defer cm.mtx.Unlock()

	switch nv := v.(type) {
	case int:
		cm.Value += counter(nv)
	default:
		if newValue, ok := nv.(int64); ok {
			cm.Value += counter(newValue)
		}
	}
}

func (cm *CounterMetric) CreateNewMetric() Metrics {
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
