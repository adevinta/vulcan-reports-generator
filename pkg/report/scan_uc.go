/*
Copyright 2021 Adevinta
*/

package report

import (
	"context"
	"time"

	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"

	"github.com/adevinta/vulcan-reports-generator/pkg/model"
	"github.com/adevinta/vulcan-reports-generator/pkg/storage"
)

const (
	// defRfshSpan is the default refresh
	// span (s) on which update the report
	// status while it's being generated.
	defRfshSpan = 10 * time.Second
)

type scanUC struct {
	log        *log.Logger
	generator  Generator
	repository storage.ReportsRepository
}

func newScanUC(logger *log.Logger, generator Generator, repository storage.ReportsRepository) GenerateUC {
	return &scanUC{
		log:        logger,
		generator:  generator,
		repository: repository,
	}
}

func (uc *scanUC) Generate(ctx context.Context, teamInfo teamInfo, reportData interface{}) (model.Report, error) {
	scanReq, err := parseScanRequest(reportData)
	if err != nil {
		return nil, err
	}

	report := &model.ScanReport{
		BaseReport: model.BaseReport{
			Status: model.StatusGenerating,
		},
		ScanID:      scanReq.ScanID,
		ProgramName: scanReq.ProgramName,
	}

	// Save initial report.
	uc.log.WithFields(log.Fields{
		"teamID": teamInfo.ID,
		"type":   "scan",
		"scanID": scanReq.ScanID,
	}).Debug("Saving initial report")
	err = uc.repository.SaveReport(ctx, report)
	if err != nil {
		return nil, err
	}

	// Generate report and refresh status
	// periodically while generating.
	rfshCtx, cancelRfsh := context.WithCancel(context.Background())
	go uc.rfshReport(rfshCtx, uc.repository, *report, defRfshSpan)

	data, err := uc.generator.Generate(ctx, teamInfo, scanReq)
	cancelRfsh()
	if err != nil {
		uc.Finish(ctx, report.ID, model.StatusFailed)
		return nil, err
	}
	scanReportData := data.(scanReportData)

	// Update report.
	uc.log.WithFields(log.Fields{
		"teamID":   teamInfo.ID,
		"reportID": report.ID,
		"type":     "scan",
		"scanID":   scanReq.ScanID,
	}).Debug("Updating report")
	report.Files = scanReportData.Files
	report.ReportURL = scanReportData.ReportURL
	report.ReportJSONURL = scanReportData.ReportJSONURL
	report.Notification.Subject = scanReportData.EmailSubject
	report.Notification.Body = scanReportData.EmailBody
	report.Notification.Fmt = model.NotifFmtHTML
	report.DeliveredTo = teamInfo.Recipients
	report.Risk = scanReportData.Risk

	err = uc.repository.SaveReport(ctx, report)
	if err != nil {
		return nil, err
	}

	return report, nil
}

func (uc *scanUC) Finish(ctx context.Context, reportID, status string) error {
	report, err := uc.repository.GetReport(ctx, reportID)
	if err != nil {
		return err
	}
	scanReport := report.(*model.ScanReport)

	uc.log.WithFields(log.Fields{
		"reportID": scanReport.ID,
		"type":     "scan",
		"scanID":   scanReport.ScanID,
		"status":   status,
	}).Info("Finishing report generation")

	scanReport.Status = status
	return uc.repository.SaveReport(ctx, scanReport)
}

func parseScanRequest(req interface{}) (scanRequest, error) {
	scanReq := scanRequest{}
	err := mapstructure.Decode(req, &scanReq)
	if err != nil || scanReq.ScanID == "" || scanReq.ProgramName == "" {
		return scanRequest{}, ErrInvalidRequest
	}
	return scanReq, nil
}

// rfshReport updates report status to 'generating' until ctx is done.
func (uc *scanUC) rfshReport(ctx context.Context, repository storage.ReportsRepository,
	r model.ScanReport, span time.Duration) {
	ticker := time.NewTicker(span)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.Status = model.StatusGenerating
			r.UpdateStatusAt = time.Now()
			repository.SaveReport(ctx, &r)
		case <-ctx.Done():
			return
		}
	}
}
