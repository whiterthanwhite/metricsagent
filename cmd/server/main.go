package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/whiterthanwhite/metricsagent/internal/handlers"
	"github.com/whiterthanwhite/metricsagent/internal/metricdb"
	"github.com/whiterthanwhite/metricsagent/internal/runtime/metrics"
	"github.com/whiterthanwhite/metricsagent/internal/settings"
	"github.com/whiterthanwhite/metricsagent/internal/storage"
)

var (
	ServerSettings    = settings.GetSysSettings()
	flagAddress       *string
	flagRestore       *bool
	flagStoreInterval *time.Duration
	flagStoreFile     *string
	flagHashKey       *string
	flagDBAddress     *string
)

func init() {
	flagAddress = flag.String("a", settings.DefaultAddress, "")
	flagRestore = flag.Bool("r", settings.DefaultRestore, "")
	flagStoreInterval = flag.Duration("i", settings.DefaultStoreInterval, "")
	flagStoreFile = flag.String("f", settings.DefaultStoreFile, "")
	flagHashKey = flag.String("k", settings.DefaultHashKey, "")
	flagDBAddress = flag.String("d", settings.DefaultDBAddress, "")
}

func startSaveMetricsOnFile(ctx context.Context, serverMetrics map[string]metrics.Metrics) {
	if ServerSettings.StoreFile == "" {
		return
	}

	processing := true
	saveTicker := time.NewTicker(ServerSettings.StoreInterval)
	for processing {
		select {
		case <-saveTicker.C:
			storage.SaveMetricsOnFile(serverMetrics, ServerSettings)
		case <-ctx.Done():
			saveTicker.Stop()
			processing = false
		}
	}
}

func createMetricTable(ctx context.Context, conn *metricdb.Connection) {
	if conn.IsConnClose() {
		return
	}

	var exists bool
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	row := conn.QueryRow(ctx, "select exists (select from information_schema.tables where table_name = 'metrics');")
	cancel()
	if err := row.Scan(&exists); err != nil {
		log.Println(err.Error())
		return
	}

	if !exists {
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		_ = conn.QueryRow(ctx, "CREATE TABLE metrics (id varchar(50) not null, type varchar(50) not null, delta int, value double precision);")
		cancel()

		log.Println("table created")
	}
}

func main() {
	log.Println("Server start")

	flag.Parse()
	if ServerSettings.Address == settings.DefaultAddress {
		ServerSettings.Address = *flagAddress
	}
	if ServerSettings.Restore == settings.DefaultRestore {
		ServerSettings.Restore = ServerSettings.Restore || *flagRestore
	}
	if ServerSettings.StoreInterval == settings.DefaultStoreInterval {
		ServerSettings.StoreInterval = *flagStoreInterval
	}
	if ServerSettings.StoreFile == settings.DefaultStoreFile {
		ServerSettings.StoreFile = *flagStoreFile
	}
	if ServerSettings.Key == settings.DefaultHashKey {
		ServerSettings.Key = *flagHashKey
	}
	if ServerSettings.MetricDBAdress == settings.DefaultDBAddress {
		ServerSettings.MetricDBAdress = *flagDBAddress
	}
	log.Println(ServerSettings)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	newServerMetrics := storage.RestoreMetricsFromFile(ServerSettings)
	oldServerMetrics := metrics.GetAllMetrics()
	defer storage.SaveMetricsOnFile(newServerMetrics, ServerSettings)
	go startSaveMetricsOnFile(ctx, newServerMetrics)

	conn := metricdb.CreateConnnection(ctx, ServerSettings.MetricDBAdress)
	defer func() {
		if err := conn.CloseConnection(ctx); err != nil {
			log.Println(err.Error())
		}
	}()
	createMetricTable(ctx, conn)

	r := chi.NewRouter()

	// TODO: Add middleware
	r.Route("/", func(r chi.Router) {
		r.Get("/", handlers.GetAllMetricsFromFile(oldServerMetrics, newServerMetrics))
		r.Route("/update", func(r chi.Router) {
			r.Post("/", handlers.UpdateMetricOnServer(newServerMetrics, ServerSettings, conn))
			r.Post("/{metricType}/{metricName}/{metricValue}",
				handlers.UpdateMetricHandler(oldServerMetrics, newServerMetrics))
		})
		r.Route("/updates", func(r chi.Router) {
			r.Post("/", handlers.UpdateMetricsOnServer(newServerMetrics, ServerSettings, conn))
		})
		r.Route("/value", func(r chi.Router) {
			r.Post("/", handlers.GetMetricFromServer(newServerMetrics, ServerSettings))
			r.Get("/{metricType}/{metricName}",
				handlers.GetMetricValueFromServer(oldServerMetrics))
		})
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", handlers.CheckDatabaseConn(conn))
		})
		// r.Post("/", handlers.GetAllMetricsFromServer(serverMetrics))
	})

	port := fmt.Sprintf(":%v", strings.Split(ServerSettings.Address, ":")[1])
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal(err)
	}
	log.Println("Server Stop")
}
