package settings

import (
	"log"

	"github.com/caarlos0/env"
)

type SysSettings struct {
	Address        string `env:"ADDRESS" envDefault:"localhost:8080"`
	PollInterval   int64  `env:"POLL_INTERVAL" envDefault:"2"`
	ReportInterval int64  `env:"REPORT_INTERVAL" envDefault:"10"`
}

func GetSysSettings() SysSettings {
	sysSettings := SysSettings{}
	if err := env.Parse(&sysSettings); err != nil {
		log.Fatal(err)
	}
	return sysSettings
}
