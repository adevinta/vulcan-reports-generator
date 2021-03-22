/*
Copyright 2021 Adevinta
*/

package api

import (
	"fmt"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const (
	api     = "/api"
	version = "/v1"

	endpointFmt = "%s%s%s"

	getReportPath      = "/reports/:type/:id"
	getReportNotifPath = "/reports/:type/:id/notification"
	sendReportPath     = "/reports/:type/:id/send"

	healthCheckEndpoint = "/healthcheck"
)

// ReportsAPI represents an API
// to interact with reports.
type ReportsAPI struct {
	ReportsService *ReportsService
	echo           *echo.Echo
}

// NewReportsAPI builds a new reports API.
func NewReportsAPI(reportsService *ReportsService) *ReportsAPI {
	return &ReportsAPI{
		ReportsService: reportsService,
		echo:           echo.New(),
	}
}

// Start starts ReportsAPI to listen on specified port.
func (a *ReportsAPI) Start(port int) error {
	// Get Report: GET /reports/{type}/{id}
	getReportEndpoint := fmt.Sprintf(endpointFmt, api, version, getReportPath)
	a.echo.GET(getReportEndpoint, a.ReportsService.GetReport)

	// Get Report's notification: GET /reports/{type}/{id}/notification
	getReportNotifEndpoint := fmt.Sprintf(endpointFmt, api, version, getReportNotifPath)
	a.echo.GET(getReportNotifEndpoint, a.ReportsService.GetReportNotification)

	// Send Report: POST /reports/{type}/{id}/send
	sendReportEndpoint := fmt.Sprintf(endpointFmt, api, version, sendReportPath)
	a.echo.POST(sendReportEndpoint, a.ReportsService.SendReport)

	// Healthcheck
	a.echo.GET(healthCheckEndpoint, a.ReportsService.HealthCheck)

	a.echo.Use(middleware.Logger())
	return a.echo.Start(fmt.Sprintf(":%d", port))
}
