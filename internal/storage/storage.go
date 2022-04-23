package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
	"github.com/whiterthanwhite/metricsagent/internal/settings"
)

var (
	StorageSettings = settings.GetSysSettings()
)

func OpenMetricFileCSV() *os.File {
	f, err := os.OpenFile("tmp.DS_Store", os.O_CREATE|os.O_RDWR, 0750)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func WriteMetricsToFile(f *os.File, cMetrics map[string]metrics.Metric) {
	err := f.Truncate(0)
	if err != nil {
		log.Fatal(err)
	}
	skip := 0
	for _, cMetric := range cMetrics {
		a := fmt.Sprintf("%v;%v;%v\n", cMetric.GetTypeName(),
			cMetric.GetName(), cMetric.GetValue())
		z, err := f.WriteAt([]byte(a), int64(skip))
		if err != nil {
			log.Fatal(err)
		}
		skip += z
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

	_, err := f.Read(fileBytes)
	if err != nil {
		log.Fatal(err)
	}
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

type metricsWriter struct {
	file    *os.File
	encoder *json.Encoder
}

func NewMetricsWriter(fileName string) (*metricsWriter, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &metricsWriter{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *metricsWriter) WriteMetrics(serverMetrics map[string]metrics.Metrics) error {
	return p.encoder.Encode(serverMetrics)
}

func (p *metricsWriter) Close() error {
	return p.file.Close()
}

type metricsReader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewMetricsReader(fileName string) (*metricsReader, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &metricsReader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *metricsReader) ReadMetrics() (map[string]metrics.Metrics, error) {
	serverMetrics := make(map[string]metrics.Metrics)
	if err := c.decoder.Decode(&serverMetrics); err != nil {
		return nil, err
	}
	return serverMetrics, nil
}

func (c *metricsReader) Close() error {
	return c.file.Close()
}
