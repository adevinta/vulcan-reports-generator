/*
Copyright 2021 Adevinta
*/

package storage

import (
	"context"
	"database/sql"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/adevinta/vulcan-reports-generator/pkg/model"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

const (
	comma = ","
)

type ScanReportsRepository struct {
	db *sql.DB
}

func newScanReportsRepository(db *sql.DB) *ScanReportsRepository {
	return &ScanReportsRepository{
		db: db,
	}
}

func (r *ScanReportsRepository) SaveReport(ctx context.Context, report model.Report) error {
	scanReport, ok := report.(*model.ScanReport)
	if !ok {
		return ErrInvalidReportData
	}

	// Check if report for scanID already exists in DB.
	dbReport, err := r.GetReportByScanID(ctx, scanReport.ScanID)
	if err != nil {
		if err == ErrReportNotFound {
			// If there's no report for scan, insert.
			return r.Insert(ctx, scanReport)
		}
		return err
	}

	// Report exists for scan, so update it.
	scanReport.ID = dbReport.ID
	scanReport.CreatedAt = time.Now()
	return r.Update(ctx, scanReport)
}

func (r *ScanReportsRepository) GetReport(ctx context.Context, reportID string) (model.Report, error) {
	dbReport, err := FindScanReport(ctx, r.db, reportID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReportNotFound
		}
		return nil, err
	}
	return toModelReport(dbReport), nil
}

func (r *ScanReportsRepository) GetReportByScanID(ctx context.Context, scanID string) (*model.ScanReport, error) {
	scanReports, err := ScanReports(qm.Where("scan_id=?", scanID)).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	if len(scanReports) == 0 {
		return nil, ErrReportNotFound
	}
	if len(scanReports) > 1 {
		// Should not get here due to DB
		// unique constraint in scan_id field.
		return nil, fmt.Errorf("multiple reports for scan_id=%s", scanID)
	}

	return toModelReport(scanReports[0]), nil
}

func (r *ScanReportsRepository) Insert(ctx context.Context, report *model.ScanReport) error {
	report.CreatedAt = time.Now()
	dbReport := toDBReport(report)

	err := dbReport.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return err
	}

	report.ID = dbReport.ID
	return nil
}

func (r *ScanReportsRepository) Update(ctx context.Context, report *model.ScanReport) error {
	report.UpdatedAt = time.Now()
	_, err := toDBReport(report).Update(ctx, r.db, boil.Infer())
	return err
}

func toModelReport(dbReport *ScanReport) *model.ScanReport {
	emailBody, _ := b64.StdEncoding.DecodeString(dbReport.EmailBody) // nolint
	return &model.ScanReport{
		BaseReport: model.BaseReport{
			ID:     dbReport.ID,
			Status: dbReport.Status,
			Notification: model.Notification{
				Subject: dbReport.EmailSubject,
				Body:    string(emailBody),
				Fmt:     model.NotifFmtHTML,
			},
			DeliveredTo: strings.Split(dbReport.DeliveredTo, comma),
			CreatedAt:   dbReport.CreatedAt,
			UpdatedAt:   dbReport.UpdatedAt,
		},
		ScanID:         dbReport.ScanID,
		ReportURL:      dbReport.Report,
		ReportJSONURL:  dbReport.ReportJSON,
		ProgramName:    dbReport.ProgramName,
		Risk:           dbReport.Risk,
		UpdateStatusAt: dbReport.UpdateStatusAt,
	}
}

func toDBReport(modelReport *model.ScanReport) *ScanReport {
	return &ScanReport{
		ID:           modelReport.ID,
		ScanID:       modelReport.ScanID,
		Report:       modelReport.ReportURL,
		ReportJSON:   modelReport.ReportJSONURL,
		EmailSubject: modelReport.Notification.Subject,
		// Encode email body to b64 to comply with old versions.
		EmailBody:      b64.StdEncoding.EncodeToString([]byte(modelReport.Notification.Body)),
		DeliveredTo:    strings.Join(modelReport.DeliveredTo[:], comma),
		Status:         modelReport.Status,
		CreatedAt:      modelReport.CreatedAt,
		UpdatedAt:      modelReport.UpdatedAt,
		ProgramName:    modelReport.ProgramName,
		Risk:           modelReport.Risk,
		UpdateStatusAt: modelReport.UpdateStatusAt,
	}
}
