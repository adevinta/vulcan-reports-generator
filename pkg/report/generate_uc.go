/*
Copyright 2021 Adevinta
*/

package report

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/adevinta/vulcan-reports-generator/pkg/model"
	"github.com/adevinta/vulcan-reports-generator/pkg/storage"
)

// GenerateUC represents the Use Case interface for a report generation.
type GenerateUC interface {
	// Generate generates the report based on request data.
	Generate(ctx context.Context, teamInfo teamInfo, reportData interface{}) (model.Report, error)
	// Finish finishes report generation updating its status.
	Finish(ctx context.Context, reportID, status string) error
}

// NewGenerateUC creates a new report generate use case based on specified type.
func NewGenerateUC(typ model.ReportType, logger *log.Logger, generator Generator, repository storage.ReportsRepository) (GenerateUC, error) {
	switch typ {
	case model.ScanType:
		return &scanUC{
			log:        logger,
			generator:  generator,
			repository: repository,
		}, nil
	case model.LiveReportType:
		return &livereportUC{
			log:        logger,
			generator:  generator,
			repository: repository,
		}, nil
	default:
		return nil, ErrUnsupportedReportType
	}
}
