package metrics

type UpdateType uint8

const (
	WithoutUpdate UpdateType = 0
	RandomValue
	Additional
)

type MetricDescription struct {
	MName      string
	MType      metrictype
	UpdateType UpdateType
}

func GetStandardMetrics() map[string]MetricDescription {
	metricDescriptions := make(map[string]MetricDescription)
	metricDescriptions["Alloc"] = MetricDescription{MName: "Alloc", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["BuckHashSys"] = MetricDescription{MName: "BuckHashSys", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["Frees"] = MetricDescription{MName: "Frees", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["GCCPUFraction"] = MetricDescription{MName: "GCCPUFraction", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["OtherSys"] = MetricDescription{MName: "OtherSys", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["GCSys"] = MetricDescription{MName: "GCSys", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["HeapAlloc"] = MetricDescription{MName: "HeapAlloc", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["HeapIdle"] = MetricDescription{MName: "HeapIdle", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["HeapInuse"] = MetricDescription{MName: "HeapInuse", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["HeapObjects"] = MetricDescription{MName: "HeapObjects", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["HeapReleased"] = MetricDescription{MName: "HeapReleased", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["HeapSys"] = MetricDescription{MName: "HeapSys", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["LastGC"] = MetricDescription{MName: "LastGC", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["Lookups"] = MetricDescription{MName: "Lookups", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["MCacheInuse"] = MetricDescription{MName: "MCacheInuse", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["MCacheSys"] = MetricDescription{MName: "MCacheSys", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["MSpanInuse"] = MetricDescription{MName: "MSpanInuse", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["MSpanSys"] = MetricDescription{MName: "MSpanSys", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["Mallocs"] = MetricDescription{MName: "Mallocs", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["NextGC"] = MetricDescription{MName: "NextGC", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["NumForcedGC"] = MetricDescription{MName: "NumForcedGC", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["NumGC"] = MetricDescription{MName: "NumGC", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["PauseTotalNs"] = MetricDescription{MName: "PauseTotalNs", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["StackInuse"] = MetricDescription{MName: "StackInuse", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["StackSys"] = MetricDescription{MName: "StackSys", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["Sys"] = MetricDescription{MName: "Sys", MType: GaugeType, UpdateType: RandomValue}
	metricDescriptions["TotalAlloc"] = MetricDescription{MName: "TotalAlloc", MType: GaugeType, UpdateType: RandomValue}

	metricDescriptions["TotalMemory"] = MetricDescription{MName: "TotalMemory", MType: GaugeType, UpdateType: Additional}
	metricDescriptions["FreeMemory"] = MetricDescription{MName: "FreeMemory", MType: GaugeType, UpdateType: Additional}
	metricDescriptions["CPUutilization1"] = MetricDescription{MName: "CPUutilization1", MType: GaugeType, UpdateType: Additional}

	metricDescriptions["RandomValue"] = MetricDescription{MName: "RandomValue", MType: GaugeType, UpdateType: WithoutUpdate}
	metricDescriptions["PollCount"] = MetricDescription{MName: "PollCount", MType: CounterType, UpdateType: WithoutUpdate}
	return metricDescriptions
}
