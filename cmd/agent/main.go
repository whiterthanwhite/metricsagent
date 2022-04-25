package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
	"github.com/whiterthanwhite/metricsagent/internal/settings"
)

var (
	AgentSettings      = settings.GetSysSettings()
	flagAddress        *string
	flagReportInterval *time.Duration
	flagPollInterval   *time.Duration
	flagHashKey        *string
)

func init() {
	flagAddress = flag.String("a", settings.DefaultAddress, "")
	flagReportInterval = flag.Duration("r", settings.DefaultReportInterval, "")
	flagPollInterval = flag.Duration("p", settings.DefaultPollInterval, "")
	flagHashKey = flag.String("k", settings.DefaultHashKey, "")
}

func sendNewUpdate(agentClient *http.Client, m *metrics.Metrics) {
	bNewM, err := json.Marshal(*m)
	if err != nil {
		log.Println(err)
	}

	urlString := fmt.Sprintf("http://%s/update", AgentSettings.Address)
	requestBody := bytes.NewBuffer(bNewM)
	agentRequest, err := http.NewRequest(http.MethodPost, urlString, requestBody)
	if err != nil {
		log.Fatal(err)
	}
	agentRequest.Close = true
	agentRequest.Header.Set("Content-Type", "application/json")
	agentRequest.Header.Set("Content-Length", fmt.Sprint(requestBody.Len()))

	resp, err := agentClient.Do(agentRequest)
	if err != nil {
		log.Println(err)
	}
	if resp != nil {
		defer resp.Body.Close()
		_, err := io.Copy(io.Discard, resp.Body)
		if err != nil {
			log.Println(err)
		}
	}
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
		stringURL = fmt.Sprintf("http://%s/update/%s/%s/%v",
			AgentSettings.Address,
			m.GetTypeName(),
			m.GetName(),
			v)
	case float64:
		stringURL = fmt.Sprintf("http://%s/update/%s/%s/%g",
			AgentSettings.Address,
			m.GetTypeName(),
			m.GetName(),
			v)
	}

	return stringURL
}

func setUpHTTPClient(agentClient *http.Client) {
	agentClient.Timeout = 0 * time.Second
}

func enableTerminationSignals() {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(
		signalChannel,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGINT)
	exitChan := make(chan int)
	go func() {
		for {
			s := <-signalChannel
			switch s {
			case syscall.SIGTERM:
				log.Println("Signal terminte triggered.")
				exitChan <- 0
			case syscall.SIGQUIT:
				log.Println("Signal quit triggered.")
				exitChan <- 0
			case syscall.SIGINT:
				log.Println("Signal interrupt triggered.")
			}
		}
	}()
	exitCode := <-exitChan
	os.Exit(exitCode)
}

func main() {
	log.Println("Start Metric Agent")

	flag.Parse()
	if AgentSettings.Address == settings.DefaultAddress {
		AgentSettings.Address = *flagAddress
	}
	if AgentSettings.PollInterval == settings.DefaultPollInterval {
		AgentSettings.PollInterval = *flagPollInterval
	}
	if AgentSettings.ReportInterval == settings.DefaultReportInterval {
		AgentSettings.ReportInterval = *flagReportInterval
	}
	if AgentSettings.Key == settings.DefaultHashKey {
		AgentSettings.Key = *flagHashKey
	}
	log.Println(AgentSettings)

	go enableTerminationSignals()

	httpClient := &http.Client{}
	setUpHTTPClient(httpClient)

	addedMetrics := metrics.GetAllMetrics()

	pollTicker := time.NewTicker(AgentSettings.PollInterval)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(AgentSettings.ReportInterval)
	defer reportTicker.Stop()
	endTimer := time.NewTimer(1 * time.Minute)
	defer endTimer.Stop()

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
			pollCount.UpdateValue(counter)
			addedMetrics["PollCount"] = pollCount
		case <-reportTicker.C:
			log.Println("Send Metrics To Server")
			for _, metric := range addedMetrics {
				// old
				// sendOldUpdate(httpClient, &metric)
				// new
				newMetric := metric.CreateNewMetric()
				newMetric.Hash = newMetric.GenerateHash(AgentSettings.Key)
				sendNewUpdate(httpClient, &newMetric)
			}
		case <-endTimer.C:
			os.Exit(0)
		}
	}
}
