package main

import (
	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
	"fmt"
	// "math/rand"
	"time"
	"net/http"
	"net/url"
	"log"
	"strings"
	"io"
)

const (
	pollInterval = 2
	reportInterval = 10
	adress = "127.0.0.1"
	port = "8080"
)

func getMetricUrl(metricName string, metricType string) *url.URL {
	urlString := getMetricUrlString(metricName, metricType)
	urlMetric, err := url.Parse(urlString)
	if err != nil {
		log.Fatal(err)
	}
	return urlMetric
}

func getMetricUrlString(metricName string, metricType string) string {
	var metricValue float64 = 0.0
	var stringUrl string = fmt.Sprintf("http://%s:%s/update/%s/%s/%g",
		adress,
		port,
		metricType,
		metricName,
		metricValue)
	return stringUrl
}

func main() {
	httpClient := http.Client{}
	metricsMap := metrics.GetAllMetrics()	
	pollTicker := time.NewTicker(pollInterval * time.Second)
	reportTicker := time.NewTicker(reportInterval * time.Second)
	for {
		select {
		case <-pollTicker.C:
			for metricName := range metricsMap {
				metrics.UpdateMetric(metricName, 1)
			}
			fmt.Println("Updated")
		case <-reportTicker.C:
			for metricName := range metricsMap {
				urlMetric := getMetricUrl(metricName, "gauge")	
				httpClient.Post(urlMetric.String(),
					"text/plain",
					io.LimitReader(strings.NewReader(""), 0))
				fmt.Println(urlMetric.String())
			}
		}
	}
	
	/*
	updateTicker := time.NewTicker(pollInterval * time.Second)
	for {
		updateTime := <-updateTicker.C
		fmt.Println("updateTime", updateTime.String())
		url, err := url.Parse(stringUrl)
		if err != nil {
			log.Fatal(err)
		} 
		request, err := http.NewRequest(http.MethodPost, url.String(), nil)
		if err != nil {
			log.Fatal(err)	
		}
		request.Header.Set("Content-Type", "text/plain")
	}
	reportTicker = time.NewTicker(reportInterval * time.Second)
	for {
		reportTime := <-reportTicker.C
		fmt.Println("reportTime", reportTime.String())
	}
	*/
}
