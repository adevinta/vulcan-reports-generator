/*
Copyright 2021 Adevinta
*/

package report

import (
	"context"
	"database/sql"
	"errors"

	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"

	"github.com/adevinta/vulcan-reports-generator/pkg/model"
)

var (
	// ErrInvalidGeneratorType indicates that the specified generator type is invalid.
	ErrInvalidGeneratorType = errors.New("Invalid generator type")
	// ErrInvalidConfiguration indicates that the supplied configuration is invalid.
	ErrInvalidConfiguration = errors.New("Invalid configuration")
)

// Generator represents the interface
// for a report generator.
type Generator interface {
	Generate(ctx context.Context, teamInfo teamInfo, reportData interface{}) (interface{}, error)
}

// NewGenerator builds and returns a new generator for the specified type.
func NewGenerator(typ string, config interface{}, log *log.Logger, db *sql.DB) (Generator, error) {
	var generator Generator

	switch typ {
	case model.ScanType:
		cfg := scanGeneratorCfg{}
		err := mapstructure.Decode(config, &cfg)
		if err != nil {
			return nil, ErrInvalidConfiguration
		}
		generator = newScanGenerator(cfg, log, newScanPrinter(cfg.PrinterCfgFile))
	case model.LiveReportType:
		cfg := liveReportGeneratorCfg{}
		err := mapstructure.Decode(config, &cfg)
		if err != nil {
			return nil, ErrInvalidConfiguration
		}
		generator, err = newLiveReportGenerator(cfg, log)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrInvalidGeneratorType
	}

	return generator, nil
}
