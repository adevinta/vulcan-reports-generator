/*
Copyright 2021 Adevinta
*/

package api

import (
	"errors"
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
