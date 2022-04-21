package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func sendNewUpdate(agentClient *http.Client, m *metrics.Metrics) {
	bNewM, err := json.Marshal(*m)
	if err != nil {
		log.Println(err)
	}

	urlString := fmt.Sprintf("http://%s:%s/update", adress, port)
	requestBody := bytes.NewBuffer(bNewM)
	agentRequest, err := http.NewRequest(http.MethodPost, urlString, requestBody)
	if err != nil {
		log.Fatal(err)
	}
	agentRequest.Header.Set("Content-type", "application/json")
	resp2, err := agentClient.Do(agentRequest)
	if err != nil {
		log.Fatal(err)
	}

	var responseMetric metrics.Metrics
	if err := json.NewDecoder(resp2.Body).Decode(&responseMetric); err != nil {
		log.Println(err)
	}
	log.Println(responseMetric)
	if err := resp2.Body.Close(); err != nil {
		log.Fatal(err)
	}
	log.Println("new sended")
}

func sendOldUpdate(agentClient *http.Client, m *metrics.Metric) {
	urlMetric := getMetricURL(*m)
	resp, err := agentClient.Post(urlMetric.String(), "text/plain", nil)
	if err != nil {
		log.Println(err)
	}
	if err := resp.Body.Close(); err != nil {
		log.Fatal(err)
	}
	log.Println("old sended")
}

func getMetricURL(m metrics.Metric) *url.URL {
	urlString := getMetricURLString(m)
	urlMetric, err := url.Parse(urlString)
	if err != nil {
		log.Fatal(err)
	}
	return urlMetric
}

func getMetricURLString(m metrics.Metric) string {
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

	switch v := oldM.GetValue().(type) {
	case int64:
		newM.Delta = &v
	case float64:
		newM.Value = &v
	}

	return newM
}

func main() {
	log.Println("Start Metric Agent")
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
			randomValue.UpdateValue(float64(randomiser.Int63() % 10000))
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
			case int64:
				counter += v
			}
			pollCount.UpdateValue(counter)
			addedMetrics["PollCount"] = pollCount
		case <-reportTicker.C:
			log.Println("Send Metrics To Server")
			for _, metric := range addedMetrics {
				log.Println(metric, metric.GetValue())

				// sendOldUpdate(&httpClient, &metric) // old
				newMetric := createNewNetric(metric) // new
				log.Println(newMetric)
				// sendNewUpdate(&httpClient, &newMetric)

				urlString := fmt.Sprintf("http://%v:%v/update", adress, port)
				serverURL, err := url.Parse(urlString)
				if err != nil {
					log.Fatal(err)
				}
				resp, err := http.Post(serverURL.String(), "application/json", nil)
				if err != nil {
					log.Fatal(err)
				}
				responseBody, err := ioutil.ReadAll(resp.Body)
				if err := resp.Body.Close(); err != nil {
					log.Fatal(err)
				}
				log.Println(string(responseBody))
			}
		}
	}
}
