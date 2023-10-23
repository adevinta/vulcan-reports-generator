/*
Copyright 2021 Adevinta
*/

package storage

import (
	"context"
	"database/sql"

	"github.com/friendsofgo/errors"

	"github.com/adevinta/vulcan-reports-generator/pkg/model"
)

const (
	comma = ","
)

var (
	// ErrInvalidRepositoryType indicates that the specified repository type is not valid.
	ErrInvalidRepositoryType = errors.New("Invalid repository type")
	// ErrInvalidReportData indicates that given report data is not valid.
	ErrInvalidReportData = errors.New("Invalid report data")
	// ErrReportNotFound indicates that the specified report was not found.
	ErrReportNotFound = errors.New("Report not found")
)

// ReportsRepository represents the abstraction
// for a generic report repository.
type ReportsRepository interface {
	GetReport(ctx context.Context, reportID string) (model.Report, error)
	SaveReport(ctx context.Context, report model.Report) error
}

// NewReportsRepository builds and returns a new ReportsRepository
// for the specified type.
func NewReportsRepository(typ string, db *sql.DB) (ReportsRepository, error) {
	switch typ {
	case model.LiveReportType:
		return newLiveReportsRepository(db), nil
	default:
		return nil, ErrInvalidRepositoryType
	}
}
