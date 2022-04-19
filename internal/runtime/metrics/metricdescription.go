package metrics

type MetricDescription struct {
	MName string
	MType metrictype
}

func GetStandardMetrics() map[string]MetricDescription {
	metricDescriptions := make(map[string]MetricDescription)
	metricDescriptions["Alloc"] = MetricDescription{MName: "Alloc", MType: GaugeType}
	metricDescriptions["BuckHashSys"] = MetricDescription{MName: "BuckHashSys", MType: GaugeType}
	metricDescriptions["Frees"] = MetricDescription{MName: "Frees", MType: GaugeType}
	metricDescriptions["GCCPUFraction"] = MetricDescription{MName: "GCCPUFraction", MType: GaugeType}
	metricDescriptions["OtherSys"] = MetricDescription{MName: "OtherSys", MType: GaugeType}
	metricDescriptions["GCSys"] = MetricDescription{MName: "GCSys", MType: GaugeType}
	metricDescriptions["HeapAlloc"] = MetricDescription{MName: "HeapAlloc", MType: GaugeType}
	metricDescriptions["HeapIdle"] = MetricDescription{MName: "HeapIdle", MType: GaugeType}
	metricDescriptions["HeapInuse"] = MetricDescription{MName: "HeapInuse", MType: GaugeType}
	metricDescriptions["HeapObjects"] = MetricDescription{MName: "HeapObjects", MType: GaugeType}
	metricDescriptions["HeapReleased"] = MetricDescription{MName: "HeapReleased", MType: GaugeType}
	metricDescriptions["HeapSys"] = MetricDescription{MName: "HeapSys", MType: GaugeType}
	metricDescriptions["LastGC"] = MetricDescription{MName: "LastGC", MType: GaugeType}
	metricDescriptions["Lookups"] = MetricDescription{MName: "Lookups", MType: GaugeType}
	metricDescriptions["MCacheInuse"] = MetricDescription{MName: "MCacheInuse", MType: GaugeType}
	metricDescriptions["MCacheSys"] = MetricDescription{MName: "MCacheSys", MType: GaugeType}
	metricDescriptions["MSpanInuse"] = MetricDescription{MName: "MSpanInuse", MType: GaugeType}
	metricDescriptions["MSpanSys"] = MetricDescription{MName: "MSpanSys", MType: GaugeType}
	metricDescriptions["Mallocs"] = MetricDescription{MName: "Mallocs", MType: GaugeType}
	metricDescriptions["NextGC"] = MetricDescription{MName: "NextGC", MType: GaugeType}
	metricDescriptions["NumForcedGC"] = MetricDescription{MName: "NumForcedGC", MType: GaugeType}
	metricDescriptions["NumGC"] = MetricDescription{MName: "NumGC", MType: GaugeType}
	metricDescriptions["PauseTotalNs"] = MetricDescription{MName: "PauseTotalNs", MType: GaugeType}
	metricDescriptions["StackInuse"] = MetricDescription{MName: "StackInuse", MType: GaugeType}
	metricDescriptions["StackSys"] = MetricDescription{MName: "StackSys", MType: GaugeType}
	metricDescriptions["Sys"] = MetricDescription{MName: "Sys", MType: GaugeType}
	metricDescriptions["TotalAlloc"] = MetricDescription{MName: "TotalAlloc", MType: GaugeType}
	metricDescriptions["RandomValue"] = MetricDescription{MName: "RandomValue", MType: GaugeType}
	return metricDescriptions
}
