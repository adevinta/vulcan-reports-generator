/*
Copyright 2021 Adevinta
*/

package api

import (
	"errors"
	"strings"
	"time"

	"github.com/adevinta/vulcan-reports-generator/pkg/model"
)

var (
	ErrInvalidReportFormat = errors.New("Invalid Report Format")
)

// SendReportReqDTO represents the DTO
// for the Send Report endpoint payload.
type SendReportReqDTO struct {
	Recipients []string `json:"recipients"`
}

// ReportNotificationDTO represents the response DTO
// for the Get Report's Notification endpoint.
type ReportNotificationDTO struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Format  string `json:"format"`
}

// ScanReportDTO represents the DTO
// to return for Scan Reports requests.
type ScanReportDTO struct {
	ID            string `json:"id"`
	ReportURL     string `json:"report_url"`
	ReportJSONURL string `json:"report_json_url"`
	ScanID        string `json:"scan_id"`
	ProgramName   string `json:"program_name"`
	Status        string `json:"status"`
	Risk          int    `json:"risk"`
	EmailSubject  string `json:"email_subject"`
	EmailBody     string `json:"email_body"`
	DeliveredTo   string `json:"delivered_to"`
	// DeliveredTo is a comma separeted list
	// of all recipients to comply with previous
	// vulcan API data format.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toReportDTO(typ model.ReportType, report model.Report) (interface{}, error) {
	switch typ {
	case model.ScanType:
		return toScanReportDTO(report)
	default:
		// Should never get here because
		// type validation is done in service.
		return nil, errors.New("Unsupported Report Type")
	}
}

func toScanReportDTO(report model.Report) (ScanReportDTO, error) {
	r, ok := report.(*model.ScanReport)
	if !ok {
		return ScanReportDTO{}, ErrInvalidReportFormat
	}

	return ScanReportDTO{
		ID:            r.ID,
		ReportURL:     r.ReportURL,
		ReportJSONURL: r.ReportJSONURL,
		ScanID:        r.ScanID,
		ProgramName:   r.ProgramName,
		Status:        r.Status,
		Risk:          r.Risk,
		EmailSubject:  r.Notification.Subject,
		EmailBody:     r.Notification.Body,
		DeliveredTo:   strings.Join(r.DeliveredTo, ","),
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}, nil
}
