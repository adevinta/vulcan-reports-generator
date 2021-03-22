/*
Copyright 2021 Adevinta
*/

package report

import (
	"context"

	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"

	"github.com/adevinta/vulcan-reports-generator/pkg/model"
	"github.com/adevinta/vulcan-reports-generator/pkg/storage"
)

type livereportUC struct {
	log        *log.Logger
	generator  Generator
	repository storage.ReportsRepository
}

func newLiveReportUC(logger *log.Logger, generator Generator, repository storage.ReportsRepository) GenerateUC {
	return &livereportUC{
		log:        logger,
		generator:  generator,
		repository: repository,
	}
}

func (uc *livereportUC) Generate(ctx context.Context, teamInfo teamInfo, reportData interface{}) (model.Report, error) {
	liveReportReq, err := parseLiveReportRequest(reportData)
	if err != nil {
		return nil, err
	}

	report := &model.LiveReport{
		BaseReport: model.BaseReport{
			Status: model.StatusGenerating,
		},
		TeamID:   liveReportReq.TeamID,
		DateFrom: liveReportReq.DateFrom,
		DateTo:   liveReportReq.DateTo,
	}

	// Save initial report.
	uc.log.WithFields(log.Fields{
		"teamID":   teamInfo.ID,
		"type":     "livereport",
		"Team":     liveReportReq.TeamID,
		"DateFrom": liveReportReq.DateFrom,
		"DateTo":   liveReportReq.DateTo,
	}).Debug("Saving initial report")
	err = uc.repository.SaveReport(ctx, report)
	if err != nil {
		return nil, err
	}

	data, err := uc.generator.Generate(ctx, teamInfo, liveReportReq)
	if err != nil {
		uc.Finish(ctx, report.ID, model.StatusFailed)
		return nil, err
	}
	liveReportData := data.(liveReportData)

	// Update report.
	uc.log.WithFields(log.Fields{
		"Team":     liveReportReq.TeamID,
		"DateFrom": liveReportReq.DateFrom,
		"DateTo":   liveReportReq.DateTo,
		"reportID": report.ID,
		"type":     "livereport",
	}).Debug("Updating report")
	report.Notification.Subject = liveReportData.EmailSubject
	report.Notification.Body = liveReportData.EmailBody
	report.Notification.Fmt = model.NotifFmtHTML
	report.DeliveredTo = teamInfo.Recipients

	err = uc.repository.SaveReport(ctx, report)
	if err != nil {
		return nil, err
	}

	return report, nil
}

func (uc *livereportUC) Finish(ctx context.Context, reportID, status string) error {
	report, err := uc.repository.GetReport(ctx, reportID)
	if err != nil {
		return err
	}
	liveReport := report.(*model.LiveReport)

	uc.log.WithFields(log.Fields{
		"reportID": liveReport.ID,
		"type":     "livereport",
		"TeamID":   liveReport.TeamID,
		"DateFrom": liveReport.DateFrom,
		"DateTo":   liveReport.DateTo,
		"status":   status,
	}).Info("Finishing report generation")

	liveReport.Status = status
	return uc.repository.SaveReport(ctx, liveReport)
}

func parseLiveReportRequest(req interface{}) (liveReportRequest, error) {
	liveReportReq := liveReportRequest{}
	err := mapstructure.Decode(req, &liveReportReq)
	if err != nil || liveReportReq.TeamID == "" || liveReportReq.DateFrom == "" || liveReportReq.DateTo == "" {
		return liveReportRequest{}, ErrInvalidRequest
	}
	return liveReportReq, nil
}
