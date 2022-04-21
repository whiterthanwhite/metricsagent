package settings

import (
	"log"
	"os"
	"strconv"
)

const (
	defaultPollInterval   = "2"
	defaultReportInterval = "10"
	defaultAddress        = "localhost:8080"
)

type SysSettings struct {
	Address        string
	PollInterval   int64
	ReportInterval int64
}

func GetSysSettings() SysSettings {
	sysSettings := SysSettings{}
	sysSettings.Address = getEnvVariable("ADDRESS")
	var err error
	v := getEnvVariable("REPORT_INTERVAL")
	sysSettings.ReportInterval, err = strconv.ParseInt(v, 0, 8)
	if err != nil {
		log.Fatal(err)
	}
	v = getEnvVariable("POLL_INTERVAL")
	sysSettings.PollInterval, err = strconv.ParseInt(v, 0, 8)
	if err != nil {
		log.Fatal(err)
	}
	return sysSettings
}

func getEnvVariable(envName string) string {
	v := os.Getenv(envName)
	if v == "" {
		defaultValue := ""
		switch envName {
		case "ADDRESS":
			defaultValue = defaultAddress
		case "REPORT_INTERVAL":
			defaultValue = defaultReportInterval
		case "POLL_INTERVAL":
			defaultValue = defaultPollInterval
		}

		if err := os.Setenv(envName, defaultValue); err != nil {
			log.Fatal(err)
		}
		v = defaultValue
	}
	return v
}
