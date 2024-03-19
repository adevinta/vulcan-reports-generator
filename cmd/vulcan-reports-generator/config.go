/*
Copyright 2021 Adevinta
*/

package main

import (
	"io"
	"os"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"

	"github.com/adevinta/vulcan-reports-generator/pkg/notify"
	"github.com/adevinta/vulcan-reports-generator/pkg/queue"
)

type config struct {
	Log        logConfig
	API        apiConfig
	DB         dbConfig
	SQS        sqsConfig
	SES        notify.SESConfig
	Generators map[string]interface{}
}

type logConfig struct {
	Level     string
	AddCaller bool `toml:"add_caller"`
}

type apiConfig struct {
	Port int `toml:"port"`
}

type dbConfig struct {
	Dialect string
	Host    string `toml:"host"`
	Port    string `toml:"port"`
	SSLMode string `toml:"sslmode"`
	User    string `toml:"user"`
	Pass    string `toml:"password"`
	Name    string `toml:"name"`
}

type sqsConfig struct {
	queue.SQSConfig
	NProcessors uint8 `toml:"number_of_processors"`
}

func parseConfig(cfgFilePath string) (*config, error) {
	cfgFile, err := os.Open(cfgFilePath)
	if err != nil {
		return nil, err
	}
	defer cfgFile.Close()

	cfgData, err := io.ReadAll(cfgFile)
	if err != nil {
		return nil, err
	}

	var conf config
	if _, err := toml.Decode(string(cfgData[:]), &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

// parseLogLevel parses a configured string log level
// and returns the correspondent logrus log level.
// If log level is invalid, default level is Info.
func parseLogLevel(logLevel string) log.Level {
	switch logLevel {
	case "panic":
		return log.PanicLevel
	case "fatal":
		return log.FatalLevel
	case "error":
		return log.ErrorLevel
	case "warn":
		return log.WarnLevel
	case "info":
		return log.InfoLevel
	case "debug":
		return log.DebugLevel
	case "trace":
		return log.TraceLevel
	default:
		return log.InfoLevel
	}
}
