/*
Copyright 2021 Adevinta
*/

package report

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	metrics "github.com/adevinta/vulcan-metrics-client"
	log "github.com/sirupsen/logrus"

	"github.com/adevinta/vulcan-reports-generator/pkg/model"
	"github.com/adevinta/vulcan-reports-generator/pkg/notify"
	"github.com/adevinta/vulcan-reports-generator/pkg/queue"
)

var (
	// ErrInvalidRequest indicates that the request format is invalid.
	ErrInvalidRequest = errors.New("Invalid request")
	// ErrUnsupportedReportType indicates that the specified report type is not supported.
	ErrUnsupportedReportType = errors.New("The requested report type is not supported")
)

// genRequest represents the expected
// request for reports processor.
//
//   - Typ identifies the generator type.
//   - TeamInfo contains information related
//     to the addressee team for the report.
//   - Data represents a struct to be
//     parsed by the specified generator.
//   - AutoSend indicates if report notification
//     must be sent automatically.
type genRequest struct {
	Typ      model.ReportType `json:"type"`
	TeamInfo teamInfo         `json:"team_info"`
	Data     interface{}      `json:"data"`
	AutoSend bool             `json:"auto_send"`
}

type teamInfo struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Recipients []string `json:"recipients"`
}

type reportsProcessor struct {
	log           *log.Logger
	generateUCC   map[model.ReportType]GenerateUC
	notifier      notify.Notifier
	metricsClient metrics.Client
}

// NewProcessor builds and returns a new Reports Processor.
func NewProcessor(log *log.Logger, generateUCC map[model.ReportType]GenerateUC,
	notifier notify.Notifier, metricsClient metrics.Client) (queue.Processor, error) {
	return &reportsProcessor{
		log:           log,
		generateUCC:   generateUCC,
		notifier:      notifier,
		metricsClient: metricsClient,
	}, nil
}

// ProcessMessage processes a report generation request read
// from the queue.
func (p *reportsProcessor) ProcessMessage(mssg string) error {
	req, err := parseGenRequest(mssg)
	if err != nil {
		return err
	}
	ctx := context.Background()

	p.log.WithFields(log.Fields{
		"teamID":   req.TeamInfo.ID,
		"teamName": req.TeamInfo.Name,
		"type":     req.Typ,
		"send":     req.AutoSend,
	}).Info("Processing report")

	generateUC, ok := p.generateUCC[req.Typ]
	if !ok {
		return ErrUnsupportedReportType
	}

	// Generate.
	report, err := generateUC.Generate(ctx, req.TeamInfo, req.Data)
	if err != nil {
		return err
	}

	p.pushGenMetric(req.Typ)

	// Notify.
	if req.AutoSend {
		p.log.WithFields(log.Fields{
			"teamID":   req.TeamInfo.ID,
			"type":     req.Typ,
			"reportID": report.GetID(),
		}).Debug("Sending notification")
		notif := report.GetNotification()
		err = p.notifier.Notify(notif.Subject, notif.Body, notif.Fmt, req.TeamInfo.Recipients)
		if err != nil {
			generateUC.Finish(ctx, report.GetID(), model.StatusFailed)
			return err
		}
		p.pushNotifMetric(req.Typ)
	}

	// Set report as finished.
	return generateUC.Finish(ctx, report.GetID(), model.StatusFinished)
}

// pushGenMetric increments the number of generated reports for reportType.
func (p *reportsProcessor) pushGenMetric(reportType model.ReportType) {
	p.metricsClient.Push(metrics.Metric{
		Name:  "vulcan.report.generated",
		Typ:   metrics.Count,
		Value: 1,
		Tags:  []string{fmt.Sprint("reporttype:", reportType)},
	})
}

// pushNotifMetric increments the number of notified reports for reportType.
func (p *reportsProcessor) pushNotifMetric(reportType model.ReportType) {
	p.metricsClient.Push(metrics.Metric{
		Name:  "vulcan.report.notified",
		Typ:   metrics.Count,
		Value: 1,
		Tags:  []string{fmt.Sprint("reporttype:", reportType)},
	})
}

func parseGenRequest(reqData string) (genRequest, error) {
	// Validate generic fields.
	var req genRequest
	err := json.Unmarshal([]byte(reqData), &req)
	if err != nil {
		return genRequest{}, fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}

	if req.TeamInfo.ID == "" || req.Typ == "" {
		return genRequest{}, ErrInvalidRequest
	}

	return req, nil
}
