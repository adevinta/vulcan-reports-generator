/*
Copyright 2021 Adevinta
*/

package model

import "time"

// LiveReport represents a report
// for vulcan team.
type LiveReport struct {
	BaseReport
	TeamID   string
	DateFrom string
	DateTo   string
}

func (r *LiveReport) GetID() string {
	return r.ID
}

func (r *LiveReport) GetNotification() Notification {
	return r.Notification
}

func (r *LiveReport) GetDeliveredTo() []string {
	return r.DeliveredTo
}

func (r *LiveReport) GetStatus() string {
	return r.Status
}

func (r *LiveReport) GetCreatedAt() time.Time {
	return r.CreatedAt
}

func (r *LiveReport) GetUpdatedAt() time.Time {
	return r.UpdatedAt
}

func (r *LiveReport) GetTeamID() string {
	return r.TeamID
}

func (r *LiveReport) GetDateTo() string {
	return r.DateTo
}

func (r *LiveReport) GetDateFrom() string {
	return r.DateFrom
}
