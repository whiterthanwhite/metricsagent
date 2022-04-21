package settings

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env"
)

type SysSettings struct {
	Address        string
	PollInterval   time.Duration
	ReportInterval time.Duration
	StoreInterval  time.Duration
	StoreFile      string
	Restore        bool
}

type EnvSysSettings struct {
	Address        string `env:"ADDRESS" envDefault:"localhost:8080"`
	PollInterval   string `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval string `env:"REPORT_INTERVAL" envDefault:"10s"`
	StoreInterval  string `env:"REPORT_INTERVAL" envDefault:"300s"`
	StoreFile      string `env:"REPORT_INTERVAL" envDefault:"/tmp/devops-metrics-db.json"`
	Restore        bool   `env:"REPORT_INTERVAL" envDefault:"true"`
}

func GetSysSettings() SysSettings {
	envSysSettings := EnvSysSettings{}
	if err := env.Parse(&envSysSettings); err != nil {
		log.Fatal(err)
	}
	sysSettings := SysSettings{
		Address:   envSysSettings.Address,
		StoreFile: envSysSettings.StoreFile,
		Restore:   envSysSettings.Restore,
	}
	values := strings.Split(envSysSettings.PollInterval, "")
	sysSettings.PollInterval = parseDurationSettings(values)

	values = strings.Split(envSysSettings.ReportInterval, "")
	sysSettings.ReportInterval = parseDurationSettings(values)

	values = strings.Split(envSysSettings.StoreInterval, "")
	sysSettings.StoreInterval = parseDurationSettings(values)

	return sysSettings
}

func parseDurationSettings(values []string) time.Duration {
	strInt := ""
	for _, value := range values {
		switch value {
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			strInt += value
		}
	}
	v, err := strconv.ParseInt(strInt, 0, 64)
	if err != nil {
		log.Fatal(err)
	}
	return time.Duration(v) * time.Second
}
