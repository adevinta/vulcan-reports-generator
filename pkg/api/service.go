/*
Copyright 2021 Adevinta
*/

package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"github.com/adevinta/vulcan-reports-generator/pkg/model"
	"github.com/adevinta/vulcan-reports-generator/pkg/notify"
	"github.com/adevinta/vulcan-reports-generator/pkg/storage"
)

const (
	unSupportedReportType = "Unsupported Report Type"
	okResp                = "OK"
)

var (
	notifFmts = map[model.NotifFmt]string{
		model.NotifFmtHTML: "HTML",
		model.NotifFmtText: "text",
	}
)

// ReportsService represents the
// service layer for the reports API.
type ReportsService struct {
	log          *log.Logger
	notifier     notify.Notifier
	repositories map[model.ReportType]storage.ReportsRepository
}

// NewReportsService builds a new Reports API Service.
func NewReportsService(log *log.Logger, notifier notify.Notifier,
	repositories map[model.ReportType]storage.ReportsRepository) *ReportsService {
	return &ReportsService{
		log:          log,
		notifier:     notifier,
		repositories: repositories,
	}
}

// GetReport returns the report for the specified type and id.
func (s *ReportsService) GetReport(c echo.Context) error {
	id := c.Param("id")
	typ := model.ReportType(c.Param("type"))

	ctx := context.Background()

	r, ok := s.repositories[typ]
	if !ok {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, unSupportedReportType)
	}

	var err error
	var report model.Report
	// scan reports are requested by scan_id
	// so we have to cast repository to scan
	// type so we can use a non-interface method.
	if typ == model.ScanType {
		r := r.(*storage.ScanReportsRepository)
		report, err = r.GetReportByScanID(ctx, id)
	} else {
		report, err = r.GetReport(ctx, id)
	}

	if err != nil {
		if errors.Is(err, storage.ErrReportNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return err
	}

	respDTO, err := toReportDTO(typ, report)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, respDTO)
}

// GetReportNotification returns the report's notification data for the specified type and id.
func (s *ReportsService) GetReportNotification(c echo.Context) error {
	id := c.Param("id")
	typ := model.ReportType(c.Param("type"))

	ctx := context.Background()

	r, ok := s.repositories[typ]
	if !ok {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, unSupportedReportType)
	}

	var err error
	var report model.Report
	// scan reports are requested by scan_id
	// so we have to cast repository to scan
	// type so we can use a non-interface method.
	if typ == model.ScanType {
		r := r.(*storage.ScanReportsRepository)
		report, err = r.GetReportByScanID(ctx, id)
	} else {
		report, err = r.GetReport(ctx, id)
	}

	if err != nil {
		if errors.Is(err, storage.ErrReportNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return err
	}

	respDTO := ReportNotificationDTO{
		Subject: report.GetNotification().Subject,
		Body:    report.GetNotification().Body,
		Format:  notifFmts[report.GetNotification().Fmt],
	}

	return c.JSON(http.StatusOK, respDTO)
}

// SendReport sends the report notification for the specified report type and id.
func (s *ReportsService) SendReport(c echo.Context) error {
	id := c.Param("id")
	typ := model.ReportType(c.Param("type"))

	ctx := context.Background()

	req := SendReportReqDTO{}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}

	r, ok := s.repositories[typ]
	if !ok {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, unSupportedReportType)
	}

	var err error
	var report model.Report
	// scan reports are requested by scan_id
	// so we have to cast repository to scan
	// type so we can use a non-interface method.
	if typ == model.ScanType {
		r := r.(*storage.ScanReportsRepository)
		report, err = r.GetReportByScanID(ctx, id)
	} else {
		report, err = r.GetReport(ctx, id)
	}

	if err != nil {
		if errors.Is(err, storage.ErrReportNotFound) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return err
	}

	if report.GetStatus() == model.StatusFailed {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "report generation failed")
	}
	if report.GetStatus() == model.StatusGenerating {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "report is being generated")
	}

	notif := report.GetNotification()
	err = s.notifier.Notify(notif.Subject, notif.Body, notif.Fmt, req.Recipients)
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, okResp)
}

// HealthCheck is the service handler for healthcheck queries.
func (s *ReportsService) HealthCheck(c echo.Context) error {
	return c.String(http.StatusOK, okResp)
}
