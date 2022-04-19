package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
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
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			for _, m := range addedMetrics {
				m.UpdateValue(25.0)
			}
		case <-reportTicker.C:
			for _, m := range addedMetrics {
				urlMetric := getMetricURL(m)

				resp1, err := httpClient.Post(urlMetric.String(), "text/plain", bytes.NewBuffer([]byte{}))
				if err != nil {
					log.Fatal(err)
				}
				resp1.Body.Close()

				resp2, err := httpClient.Post(fmt.Sprintf("http://%s:%s/update/", adress, port),
					"application/json", bytes.NewBuffer([]byte{}))
				if err != nil {
					log.Fatal(err)
				}
				resp2.Body.Close()

				log.Println(resp1.Status, resp1.Header.Get("Content-Type"))
				log.Println(resp2.Status, resp2.Header.Get("Content-Type"))
			}
		}
	}
}
