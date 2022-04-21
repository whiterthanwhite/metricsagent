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
}

type EnvSysSettings struct {
	Address        string `env:"ADDRESS" envDefault:"localhost:8080"`
	PollInterval   string `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval string `env:"REPORT_INTERVAL" envDefault:"10s"`
}

func GetSysSettings() SysSettings {
	envSysSettings := EnvSysSettings{}
	if err := env.Parse(&envSysSettings); err != nil {
		log.Fatal(err)
	}
	sysSettings := SysSettings{
		Address: envSysSettings.Address,
	}
	values := strings.Split(envSysSettings.PollInterval, "")
	sysSettings.PollInterval = parseDurationSettings(values)

	values = strings.Split(envSysSettings.ReportInterval, "")
	sysSettings.ReportInterval = parseDurationSettings(values)

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
