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

type LiveReportsRepository struct {
	db *sql.DB
}

func newLiveReportsRepository(db *sql.DB) *LiveReportsRepository {
	return &LiveReportsRepository{
		db: db,
	}
}

func (r *LiveReportsRepository) SaveReport(ctx context.Context, report model.Report) error {
	liveReport, ok := report.(*model.LiveReport)
	if !ok {
		return ErrInvalidReportData
	}

	// Check if report for teamID+dateFrom+dateTo already exists in DB.
	dbReport, err := r.GetReportByTeamAndDateRange(ctx, liveReport.TeamID, liveReport.DateFrom, liveReport.DateTo)
	if err != nil {
		if err == ErrReportNotFound {
			// If there's no report, insert.
			return r.Insert(ctx, liveReport)
		}
		return err
	}

	// Report exists, so update it.
	liveReport.ID = dbReport.ID
	liveReport.CreatedAt = time.Now()
	return r.Update(ctx, liveReport)
}

func (r *LiveReportsRepository) GetReport(ctx context.Context, reportID string) (model.Report, error) {
	dbReport, err := FindLiveReport(ctx, r.db, reportID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReportNotFound
		}
		return nil, err
	}
	return toModelLiveReport(dbReport), nil
}

func (r *LiveReportsRepository) GetReportByTeamAndDateRange(ctx context.Context, teamID string, dateFrom string, dateTo string) (*model.LiveReport, error) {
	liveReports, err := LiveReports(qm.Where("team_id=? AND date_from=? AND date_to=?", teamID, dateFrom, dateTo)).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	if len(liveReports) == 0 {
		return nil, ErrReportNotFound
	}
	if len(liveReports) > 1 {
		// Should not get here due to DB
		// unique constraint in (team_id, dateFrom, dateTo) fields.
		return nil, fmt.Errorf("multiple reports for team_id=%s, date_to=%s, date_from=%s", teamID, dateFrom, dateTo)
	}

	return toModelLiveReport(liveReports[0]), nil
}

func (r *LiveReportsRepository) Insert(ctx context.Context, report *model.LiveReport) error {
	report.CreatedAt = time.Now()
	dbReport := toDBLiveReport(report)

	err := dbReport.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return err
	}

	report.ID = dbReport.ID
	return nil
}

func (r *LiveReportsRepository) Update(ctx context.Context, report *model.LiveReport) error {
	report.UpdatedAt = time.Now()
	_, err := toDBLiveReport(report).Update(ctx, r.db, boil.Infer())
	return err
}

func toModelLiveReport(dbReport *LiveReport) *model.LiveReport {
	emailBody, _ := b64.StdEncoding.DecodeString(dbReport.EmailBody) // nolint
	return &model.LiveReport{
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
		TeamID:   dbReport.TeamID,
		DateFrom: dbReport.DateFrom,
		DateTo:   dbReport.DateTo,
	}
}

func toDBLiveReport(modelReport *model.LiveReport) *LiveReport {
	return &LiveReport{
		ID:           modelReport.ID,
		TeamID:       modelReport.TeamID,
		DateFrom:     modelReport.DateFrom,
		DateTo:       modelReport.DateTo,
		EmailSubject: modelReport.Notification.Subject,
		// Encode email body to b64 to comply with old versions.
		EmailBody:   b64.StdEncoding.EncodeToString([]byte(modelReport.Notification.Body)),
		DeliveredTo: strings.Join(modelReport.DeliveredTo[:], comma),
		Status:      modelReport.Status,
		CreatedAt:   modelReport.CreatedAt,
		UpdatedAt:   modelReport.UpdatedAt,
	}
}
