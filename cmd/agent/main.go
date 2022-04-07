package main

import (
	"fmt"

	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"

	// "math/rand"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	pollInterval   = 2
	reportInterval = 10
	adress         = "127.0.0.1"
	port           = "8080"
)

func getMetricURL(m metrics.Metric) *url.URL {
	urlString := getMetricURLString(m)
	urlMetric, err := url.Parse(urlString)
	if err != nil {
		log.Fatal(err)
	}
	return urlMetric
}

func getMetricURLString(m metrics.Metric) string {
	metricValue := m.GetValue().(float64)
	stringURL := fmt.Sprintf("http://%s:%s/update/%s/%s/%g",
		adress,
		port,
		m.GetTypeName(),
		m.GetName(),
		metricValue)
	return stringURL
}

func main() {
	httpClient := http.Client{}
	addedMetrics := metrics.GetAllMetrics()
	pollTicker := time.NewTicker(pollInterval * time.Second)
	reportTicker := time.NewTicker(reportInterval * time.Second)
	for {
		select {
		case <-pollTicker.C:
			for _, m := range addedMetrics {
				m.UpdateValue(25.0)
			}
		case <-reportTicker.C:
			for _, m := range addedMetrics {
				urlMetric := getMetricURL(m)
				resp, err := httpClient.Post(urlMetric.String(), "text/plain",
					io.LimitReader(strings.NewReader(""), 0))
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(resp.Status, resp.Header.Get("Content-Type"))
				resp.Body.Close()
			}
		}
	}
}
