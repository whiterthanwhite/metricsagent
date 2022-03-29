package metrics

type (
	gauge float64
	counter int64
)

var (
	gaugeMetrics = map[string]gauge{
		"Alloc": 0,
		"BuckHashSys": 0,
		"Frees": 0,
		"GCCPUFraction": 0,
		"GCSys": 0,
		"HeapAlloc": 0,
		"HeapIdle": 0,
		"HeapInuse": 0,
		"HeapObjects": 0,
		"HeapReleased": 0,
		"HeapSys": 0,
		"LastGC": 0,
		"Lookups": 0,
		"MCacheInuse": 0,
		"MCacheSys": 0,
		"MSpanInuse": 0,
		"MSpanSys": 0,
		"Mallocs": 0,
		"NextGC": 0,
		"NumForcedGC": 0,
		"NumGC": 0,
		"OtherSys": 0,
		"PauseTotalNs": 0,
		"StackInuse": 0,
		"StackSys": 0,
		"Sys": 0,
		"TotalAlloc": 0,
	}
	counterMetrics = map[string]counter{
		"PollCount": 0,
	}
	updateMetrics = map[string]gauge{
		"RandomValue": 0,
	}
)

func GetAllMetrics() map[string]gauge{
	return gaugeMetrics
}

func UpdateMetric(name string, value gauge){
	gaugeMetrics[name] += value
	counterMetrics["PollCount"] += 1
}
