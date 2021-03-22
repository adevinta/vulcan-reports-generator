/*
Copyright 2021 Adevinta
*/

package model

import (
	"time"
)

// NotifFmt represents the
// report notification format.
type NotifFmt uint8

const (
	// NotifFmtHTML indicates notification has HTML format.
	NotifFmtHTML = iota
	// NotifFmtText indicates notification has Text format.
	NotifFmtText

	// ScanType identifies the scan report type.
	ScanType = "scan"

	// LiveReportType identifies the live report type.
	LiveReportType = "livereport"

	// StatusGenerating indicates that report is being generated.
	StatusGenerating = "GENERATING"
	// StatusFinished indicates that report has been generated.
	StatusFinished = "FINISHED"
	// StatusFailed indicates that report generation failed.
	StatusFailed = "FAILED"
)

// ReportType specifies a report type.
type ReportType string

// Report represents the interface
// which all report types must comply with.
type Report interface {
	GetID() string
	GetFiles() []FileInfo
	GetNotification() Notification
	GetDeliveredTo() []string
	GetStatus() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
}

// BaseReport represents the common
// fields for all types of reports.
type BaseReport struct {
	ID           string
	Files        []FileInfo
	Notification Notification
	DeliveredTo  []string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// FileInfo contains report file
// and its upload destination.
type FileInfo struct {
	FilePath   string
	DestBucket string
	DestPath   string
}

// Notification represents
// a report notification.
type Notification struct {
	Subject string
	Body    string
	Fmt     NotifFmt
}
