/*
Copyright 2021 Adevinta
*/

package model

import "time"

// ScanReport represents a report
// for a vulcan scan.
type ScanReport struct {
	BaseReport
	ScanID         string
	ReportURL      string
	ReportJSONURL  string
	ProgramName    string
	Risk           int
	UpdateStatusAt time.Time
}

func (r *ScanReport) GetID() string {
	return r.ID
}

func (r *ScanReport) GetFiles() []FileInfo {
	return r.Files
}

func (r *ScanReport) GetNotification() Notification {
	return r.Notification
}

func (r *ScanReport) GetDeliveredTo() []string {
	return r.DeliveredTo
}

func (r *ScanReport) GetStatus() string {
	return r.Status
}

func (r *ScanReport) GetCreatedAt() time.Time {
	return r.CreatedAt
}

func (r *ScanReport) GetUpdatedAt() time.Time {
	return r.UpdatedAt
}
