package settings

import (
	"log"
	"time"

	"github.com/caarlos0/env"
)

const (
	DefaultAddress        string        = "localhost:8080"
	DefaultReportInterval time.Duration = 10 * time.Second
	DefaultPollInterval   time.Duration = 2 * time.Second
	DefaultStoreInterval  time.Duration = 300 * time.Second
	DefaultStoreFile      string        = "/tmp/devops-metrics-db.json"
	DefaultRestore        bool          = true
)

type SysSettings struct {
	Address        string        `env:"ADDRESS" envDefault:"localhost:8080"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	StoreInterval  time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
	StoreFile      string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore        bool          `env:"RESTORE" envDefault:"true"`
}

func GetSysSettings() SysSettings {
	sysSettings := SysSettings{}
	if err := env.Parse(&sysSettings); err != nil {
		log.Fatal(err)
	}

	return sysSettings
}
