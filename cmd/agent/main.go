package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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
	case metrics.Counter:
		mDelta = int64(v)
		newM.Delta = &mDelta
	case metrics.Gauge:
		mValue = float64(v)
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
			randomValue := addedMetrics["RandomValue"]
			randomiser := rand.NewSource(time.Now().Unix())
			randomValue.UpdateValue(float64(randomiser.Int63()))
			addedMetrics["RandomValue"] = randomValue
			var counter int64 = 0
			for _, m := range addedMetrics {
				if m.GetName() != "RandomValue" && m.GetName() != "PollCount" {
					m.UpdateValue(randomValue.GetValue())
					counter++
				}
			}
			pollCount := addedMetrics["PollCount"]
			switch v := pollCount.GetValue().(type) {
			case metrics.Counter:
				counter += int64(v)
			}
			pollCount.UpdateValue(counter)
			addedMetrics["PollCount"] = pollCount
		case <-reportTicker.C:
			for _, m := range addedMetrics {
				/*
					urlMetric := getMetricURL(m)
					resp1, err := httpClient.Post(urlMetric.String(), "text/plain", bytes.NewBuffer([]byte{}))
					if err != nil {
						log.Fatal(err)
					}
					resp1.Body.Close()
				*/
				if m.GetName() == "PollCount" {
					log.Println(m, m.GetValue())
				}
				newM := createNewNetric(m)

				bNewM, err := json.Marshal(newM)
				if err != nil {
					panic(err)
				}

				if m.GetName() == "PollCount" {
					log.Println(m, m.GetValue())
					log.Println(string(bNewM))
				}

				resp2, err := httpClient.Post(fmt.Sprintf("http://%s:%s/update/", adress, port),
					"application/json", bytes.NewBuffer(bNewM))
				if err != nil {
					log.Fatal(err)
				}
				var responseMetric metrics.Metrics
				if err := json.NewDecoder(resp2.Body).Decode(&responseMetric); err != nil {
					panic(err)
				}
				resp2.Body.Close()
			}
		}
	}
}
