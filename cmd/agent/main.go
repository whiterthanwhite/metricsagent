package main

import (
	"bytes"
	"encoding/json"
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
	log.Println(m)
	stringURL := ""
	switch v := m.GetValue().(type) {
	case int64:
		stringURL = fmt.Sprintf("http://%s:%s/update/%s/%s/%v",
			adress,
			port,
			m.GetTypeName(),
			m.GetName(),
			v)
	case float64:
		stringURL = fmt.Sprintf("http://%s:%s/update/%s/%s/%g",
			adress,
			port,
			m.GetTypeName(),
			m.GetName(),
			v)
	}

	return stringURL
}

func createNewNetric(oldM metrics.Metric) metrics.Metrics {
	newM := metrics.Metrics{
		ID:    oldM.GetName(),
		MType: oldM.GetTypeName(),
	}
	var mDelta int64 = 0
	var mValue float64 = 0

	switch v := oldM.GetValue().(type) {
	case int64:
		mDelta = v
	case float64:
		mValue = v
	}
	switch newM.MType {
	case metrics.CounterType:
		newM.Delta = &mDelta
	case metrics.GaugeType:
		newM.Value = &mValue
	}

	return newM
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

				newM := createNewNetric(m)
				bNewM, err := json.Marshal(newM)
				if err != nil {
					panic(err)
				}
				resp2, err := httpClient.Post(fmt.Sprintf("http://%s:%s/update/", adress, port),
					"application/json", bytes.NewBuffer(bNewM))
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
