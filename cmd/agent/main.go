package main

import (
	"bytes"
	"context"
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

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

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

func sendMetricsToServer(agentClient *http.Client, ms []metrics.Metrics) {
	msJSON, err := json.Marshal(ms)
	if err != nil {
		log.Println(err.Error())
		return
	}

	urlString := fmt.Sprintf("http://%s/updates/", AgentSettings.Address)
	requestBodyBuff := bytes.NewBuffer(msJSON)
	request, err := http.NewRequest(http.MethodPost, urlString, requestBodyBuff)
	if err != nil {
		log.Println(err.Error())
		return
	}

	request.Close = true
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Length", fmt.Sprint(requestBodyBuff.Len()))

	response, err := agentClient.Do(request)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println(err.Error())
		return
	}

	log.Println("Response body: ", string(responseBody))
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

func enableTerminationSignals(cancel context.CancelFunc) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(
		signalChannel,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGINT)
	s := <-signalChannel
	switch s {
	case syscall.SIGTERM:
		log.Println("Signal terminate triggered.")
	case syscall.SIGQUIT:
		log.Println("Signal quit triggered.")
	case syscall.SIGINT:
		log.Println("Signal interrupt triggered.")
	}
	cancel()
}

func UpdateStandardMetrics(agentMetrics map[string]metrics.Metric, ctx context.Context) {
	processing := true
	pollTicker := time.NewTicker(AgentSettings.PollInterval)

	pollCount := agentMetrics["PollCount"]
	randomValue := agentMetrics["RandomValue"]
	for processing {
		rand.Seed(time.Now().Unix())
		randomValue.UpdateValue(rand.Float64())
		pollCount.UpdateValue(1)

		select {
		case <-pollTicker.C:
			for _, m := range agentMetrics {
				currMetric := m.GetName()
				switch currMetric {
				case "Alloc", "BuckHashSys", "Frees", "GCCPUFraction",
					"OtherSys", "GCSys", "HeapAlloc", "HeapIdle", "HeapInuse",
					"HeapObjects", "HeapReleased", "HeapSys", "LastGC",
					"Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse",
					"MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC",
					"PauseTotalNs", "StackInuse", "StackSys", "Sys",
					"TotalAlloc":
					m.UpdateValue(randomValue.GetValue())
					pollCount.UpdateValue(1)
				}
			}
		case <-ctx.Done():
			pollTicker.Stop()
			processing = false
		}
	}
}

func UpdateAdditionalMetrics(agentMetrics map[string]metrics.Metric, ctx context.Context) {
	processing := true
	pollTicker := time.NewTicker(AgentSettings.PollInterval)

	totalMemory := agentMetrics["TotalMemory"]
	freeMemory := agentMetrics["FreeMemory"]
	cpuUtilization1 := agentMetrics["CPUutilization1"]
	pollCount := agentMetrics["PollCount"]

	for processing {
		select {
		case <-pollTicker.C:
			cpuInfo, err := cpu.Times(false)
			if err != nil {
				log.Fatal(err)
			}
			memInfo, err := mem.VirtualMemory()
			if err != nil {
				log.Fatal(err)
			}

			totalMemory.UpdateValue(float64(memInfo.Total))
			freeMemory.UpdateValue(float64(memInfo.Free))
			cpuUtilization1.UpdateValue(cpuInfo[0].User + cpuInfo[0].System -
				cpuInfo[0].Idle - cpuInfo[0].Steal)
			pollCount.UpdateValue(3)
		case <-ctx.Done():
			pollTicker.Stop()
			processing = false
		}
	}
}

func MainSendFunction(agentMetrics map[string]metrics.Metric, httpClient *http.Client, ctx context.Context) {
	log.Println("Send Metrics To Server")
	processing := true
	reportTicker := time.NewTicker(AgentSettings.ReportInterval)

	for processing {
		select {
		case <-reportTicker.C:
			ms := make([]metrics.Metrics, 0)
			for _, metric := range agentMetrics {
				// sendOldUpdate(httpClient, &metric)
				newMetric := metric.CreateNewMetric()
				newMetric.Hash = newMetric.GenerateHash(AgentSettings.Key)
				// sendNewUpdate(httpClient, &newMetric)
				ms = append(ms, newMetric)
			}
			sendMetricsToServer(httpClient, ms)
		case <-ctx.Done():
			reportTicker.Stop()
			processing = false
		}
	}
}

func main() {
	log.Println("Start agent")

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

	httpClient := &http.Client{}
	setUpHTTPClient(httpClient)

	agentMetrics := metrics.GetAllMetrics()
	ctx, cancel := context.WithCancel(context.Background())

	go enableTerminationSignals(cancel)
	go UpdateStandardMetrics(agentMetrics, ctx)
	go UpdateAdditionalMetrics(agentMetrics, ctx)
	go MainSendFunction(agentMetrics, httpClient, ctx)

	<-ctx.Done()
	fmt.Println("Finish agent")
}
