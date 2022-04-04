package storage

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
)

func OpenMetricFileCSV() *os.File {
	f, err := os.OpenFile("tmp.csv", os.O_CREATE|os.O_RDWR, 0750)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func WriteMetricsToFile(f *os.File, cMetrics map[string]metrics.Metric) {
	f.Truncate(0)
	var skip int = 0
	for _, cMetric := range cMetrics {
		a := fmt.Sprintf("%v;%v;%v\n", cMetric.GetTypeName(),
			cMetric.GetName(), cMetric.GetValue())
		z, err := f.WriteAt([]byte(a), int64(skip))
		if err != nil {
			log.Fatal(err)
		}
		skip += z
		fmt.Println(a, skip)
	}
}

func GetMetricsFromFile(f *os.File) map[string]metrics.Metric {
	fi, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}
	return getMetricsFromFile(f, fi)
}

func getMetricsFromFile(f *os.File, fi os.FileInfo) map[string]metrics.Metric {
	fileMetrics := make(map[string]metrics.Metric)
	fileBytes := make([]byte, fi.Size())

	f.Read(fileBytes)
	fileText := string(fileBytes)
	metricsStrings := strings.Split(fileText, "\n")
	if len(metricsStrings) > 0 {
		for _, metricString := range metricsStrings {
			if len(metricString) > 0 {
				m := metrics.ParseCSVString(metricString)
				if m == nil {
					panic("Cannot parse metric!")
				}
				fileMetrics[m.GetName()] = m
			}
		}
	}
	return fileMetrics
}
