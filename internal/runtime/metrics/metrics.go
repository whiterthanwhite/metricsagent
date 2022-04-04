package metrics

import (
	"log"
	"strconv"
	"strings"
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
}

type CounterMetric struct {
	Name     string
	TypeName metrictype
	Value    counter
}

func (gm GaugeMetric) GetName() string {
	return gm.Name
}

func (gm GaugeMetric) GetTypeName() metrictype {
	return gm.TypeName
}

func (gm GaugeMetric) GetValue() interface{} {
	return float64(gm.Value)
}

func (gm *GaugeMetric) UpdateValue(v interface{}) {
	newValue, ok := v.(float64)
	if ok {
		gm.Value = gauge(newValue)
		counterValue := counterMetrics["PollCount"].GetValue()
		counterMetrics["PollCount"].UpdateValue(counterValue.(counter) + 1)
	}
}

func (cm CounterMetric) GetName() string {
	return cm.Name
}

func (cm CounterMetric) GetTypeName() metrictype {
	return cm.TypeName
}

func (cm CounterMetric) GetValue() interface{} {
	return counter(cm.Value)
}

func (cm *CounterMetric) UpdateValue(v interface{}) {
	newValue, ok := v.(counter)
	if ok {
		cm.Value = newValue
	}
}

var (
	metrics        = make(map[string]Metric)
	counterMetrics = make(map[string]Metric)
)

func init() {
	metrics["Alloc"] = &GaugeMetric{
		Name:     "Alloc",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["BuckHashSys"] = &GaugeMetric{
		Name:     "BuckHashSys",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["Frees"] = &GaugeMetric{
		Name:     "Frees",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["GCCPUFraction"] = &GaugeMetric{
		Name:     "GCCPUFraction",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["GCSys"] = &GaugeMetric{
		Name:     "GCSys",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["HeapAlloc"] = &GaugeMetric{
		Name:     "HeapAlloc",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["HeapIdle"] = &GaugeMetric{
		Name:     "HeapIdle",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["HeapInuse"] = &GaugeMetric{
		Name:     "HeapInuse",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["HeapObjects"] = &GaugeMetric{
		Name:     "HeapObjects",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["HeapReleased"] = &GaugeMetric{
		Name:     "HeapReleased",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["HeapSys"] = &GaugeMetric{
		Name:     "HeapSys",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["HeapSys"] = &GaugeMetric{
		Name:     "HeapSys",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["LastGC"] = &GaugeMetric{
		Name:     "LastGC",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["Lookups"] = &GaugeMetric{
		Name:     "Lookups",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["MCacheInuse"] = &GaugeMetric{
		Name:     "MCacheInuse",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["MCacheSys"] = &GaugeMetric{
		Name:     "MCacheSys",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["MSpanInuse"] = &GaugeMetric{
		Name:     "MSpanInuse",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["MSpanSys"] = &GaugeMetric{
		Name:     "MSpanSys",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["Mallocs"] = &GaugeMetric{
		Name:     "Mallocs",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["NextGC"] = &GaugeMetric{
		Name:     "NextGC",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["NumForcedGC"] = &GaugeMetric{
		Name:     "NumForcedGC",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["NumGC"] = &GaugeMetric{
		Name:     "NumGC",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["OtherSys"] = &GaugeMetric{
		Name:     "OtherSys",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["PauseTotalNs"] = &GaugeMetric{
		Name:     "PauseTotalNs",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["StackInuse"] = &GaugeMetric{
		Name:     "StackInuse",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["StackSys"] = &GaugeMetric{
		Name:     "StackSys",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["Sys"] = &GaugeMetric{
		Name:     "Sys",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["TotalAlloc"] = &GaugeMetric{
		Name:     "TotalAlloc",
		TypeName: GaugeType,
		Value:    0,
	}
	metrics["RandomValue"] = &GaugeMetric{
		Name:     "RandomValue",
		TypeName: GaugeType,
		Value:    0,
	}
	counterMetrics["PollCount"] = &CounterMetric{
		Name:     "PollCount",
		TypeName: CounterType,
		Value:    0,
	}
}

func GetAllMetrics() map[string]Metric {
	return metrics
}

func GetMetric(name string, mType string) Metric {
	m, ok := metrics[name]
	if !ok {
		mt := metrictype(mType)
		m = createMetric(name, mt)
	}
	return m
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
	}
	return nil
}
