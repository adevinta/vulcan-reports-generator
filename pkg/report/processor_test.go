/*
Copyright 2021 Adevinta
*/

package report

import (
	"context"
	"errors"
	"reflect"
	"testing"

	metrics "github.com/adevinta/vulcan-metrics-client"
	log "github.com/sirupsen/logrus"

	"github.com/adevinta/vulcan-reports-generator/pkg/model"
	"github.com/adevinta/vulcan-reports-generator/pkg/notify"
)

var (
	errMockGen    = errors.New("ErrGen")
	errMockNotify = errors.New("ErrNotify")
	errMockFinish = errors.New("ErrFinish")

	mockReport = &model.LiveReport{
		BaseReport: model.BaseReport{
			ID:     "1",
			Status: model.StatusGenerating,
			Notification: model.Notification{
				Subject: "NotifSubject",
				Body:    "NotifBody",
				Fmt:     model.NotifFmtHTML,
			},
		},
	}
)

// Generator mock.
type mockGenerateFunc func(ctx context.Context, teamInfo teamInfo, reportData interface{}) (model.Report, error)
type mockFinishFunc func(ctx context.Context, reportID, status string) error
type mockGenerateUC struct {
	GenerateUC
	mockGenerateFunc mockGenerateFunc
	mockFinishFunc   mockFinishFunc
}

func (g *mockGenerateUC) Generate(ctx context.Context, teamInfo teamInfo, reportData interface{}) (model.Report, error) {
	return g.mockGenerateFunc(ctx, teamInfo, reportData)
}

func (g *mockGenerateUC) Finish(ctx context.Context, reportID, status string) error {
	return g.mockFinishFunc(ctx, reportID, status)
}

// Notifier mock.
type mockNotifyFunc func(subject, mssg string, fmt model.NotifFmt, recipients []string) error
type mockNotifier struct {
	notify.Notifier
	mockFunc mockNotifyFunc
}

func (n *mockNotifier) Notify(subject, mssg string, fmt model.NotifFmt, recipients []string) error {
	return n.mockFunc(subject, mssg, fmt, recipients)
}

// MetricsClient mock.
type mockMetricsClient struct {
	metrics.Client
	calls int
}

func (m *mockMetricsClient) Push(metric metrics.Metric) {
	m.calls++
}

func TestProcess(t *testing.T) {
	type fields struct {
		log           *log.Logger
		generateUCC   map[model.ReportType]GenerateUC
		notifier      notify.Notifier
		metricsClient metrics.Client
	}

	log := log.New()

	testCases := []struct {
		name                string
		fields              fields
		input               string
		expectedMetricCalls int
		expectedErr         error
	}{
		{
			name: "Should return ErrInvalidRequest, missing team info",
			fields: fields{
				log:           log,
				metricsClient: &mockMetricsClient{},
			},
			input: `
			{
				"type": "scan",
			}
			`,
			expectedMetricCalls: 0,
			expectedErr:         ErrInvalidRequest,
		},
		{
			name: "Should return ErrInvalidRequest, missing type",
			fields: fields{
				log:           log,
				metricsClient: &mockMetricsClient{},
			},
			input: `
			{
				"team_info": {
					"id": "1",
					"name": "myTeam",
					"recipients": []
				},
			}
			`,
			expectedMetricCalls: 0,
			expectedErr:         ErrInvalidRequest,
		},
		{
			name: "Should return ErrUnsupportedReportType",
			fields: fields{
				log:           log,
				metricsClient: &mockMetricsClient{},
			},
			input: `
			{
				"team_info": {
					"id": "1",
					"name": "myTeam",
					"recipients": []
				},
				"data": {},
				"type": "thisIsInvalidType"
			}`,
			expectedMetricCalls: 0,
			expectedErr:         ErrUnsupportedReportType,
		},
		{
			name: "Happy path",
			fields: fields{
				log: log,
				generateUCC: map[model.ReportType]GenerateUC{
					"scan": &mockGenerateUC{
						mockGenerateFunc: func(ctx context.Context, teamInfo teamInfo, reportData interface{}) (model.Report, error) {
							// Return mock report.
							return mockReport, nil
						},
						mockFinishFunc: func(ctx context.Context, reportID, status string) error {
							// Verify input reportID matches returned mock data from Generate.
							if reportID != mockReport.ID {
								return errors.New("reportID does not match returned mock report ID")
							}
							// Verify input status is FINISHED.
							if status != model.StatusFinished {
								return errors.New("status is not set to FINISHED")
							}
							return nil
						},
					},
				},
				notifier: &mockNotifier{
					mockFunc: func(subject, mssg string, fmt model.NotifFmt, recipients []string) error {
						// Verify input notif matches returned mock data from Generate.
						if subject != mockReport.Notification.Subject ||
							mssg != mockReport.Notification.Body ||
							fmt != mockReport.Notification.Fmt {
							return errors.New("notification data does not match mock report data")
						}
						// Verify input recipients matches process input.
						if !reflect.DeepEqual(recipients, []string{"testteam@vulcan.example.com"}) {
							return errors.New("recipients do not match processor input")
						}
						return nil
					},
				},
				metricsClient: &mockMetricsClient{},
			},
			input: `
			{
				"team_info": {
					"id": "1",
					"name": "myTeam",
					"recipients": ["testteam@vulcan.example.com"]
				},
				"data": {
					"scan_id": "1",
					"program_name": "progName"
				},
				"type": "scan",
				"auto_send": true
			}`,
			expectedMetricCalls: 2,
		},
		{
			name: "Happy path without auto send",
			fields: fields{
				log: log,
				generateUCC: map[model.ReportType]GenerateUC{
					"scan": &mockGenerateUC{
						mockGenerateFunc: func(ctx context.Context, teamInfo teamInfo, reportData interface{}) (model.Report, error) {
							// Return mock report.
							return mockReport, nil
						},
						mockFinishFunc: func(ctx context.Context, reportID, status string) error {
							// Verify input reportID matches returned mock data from Generate.
							if reportID != mockReport.ID {
								return errors.New("reportID does not match returned mock report ID")
							}
							// Verify input status is FINISHED.
							if status != model.StatusFinished {
								return errors.New("status is not set to FINISHED")
							}
							return nil
						},
					},
				},
				notifier: &mockNotifier{
					mockFunc: func(subject, mssg string, fmt model.NotifFmt, recipients []string) error {
						// There should be no call to notifier.
						return errors.New("No call expected to notify, bu got one")
					},
				},
				metricsClient: &mockMetricsClient{},
			},
			input: `
			{
				"team_info": {
					"id": "1",
					"name": "myTeam",
					"recipients": ["testteam@vulcan.example.com"]
				},
				"data": {
					"scan_id": "1",
					"program_name": "progName"
				},
				"type": "scan",
				"auto_send": false
			}`,
			expectedMetricCalls: 1,
		},
		{
			name: "Should return ErrMockGen",
			fields: fields{
				log: log,
				generateUCC: map[model.ReportType]GenerateUC{
					"scan": &mockGenerateUC{
						mockGenerateFunc: func(ctx context.Context, teamInfo teamInfo, reportData interface{}) (model.Report, error) {
							// Return Err.
							return nil, errMockGen
						},
					},
				},
				metricsClient: &mockMetricsClient{},
			},
			input: `
			{
				"team_info": {
					"id": "1",
					"name": "myTeam",
					"recipients": []
				},
				"data": {
					"scan_id": "1",
					"program_name": "progName"
				},
				"type": "scan"
			}`,
			expectedMetricCalls: 0,
			expectedErr:         errMockGen,
		},
		{
			name: "Should return ErrMockNotify",
			fields: fields{
				log: log,
				generateUCC: map[model.ReportType]GenerateUC{
					"scan": &mockGenerateUC{
						mockGenerateFunc: func(ctx context.Context, teamInfo teamInfo, reportData interface{}) (model.Report, error) {
							// Return mock report.
							return mockReport, nil
						},
						mockFinishFunc: func(ctx context.Context, reportID, status string) error {
							// Verify input reportID matches returned mock data from Generate.
							if reportID != mockReport.ID {
								return errors.New("reportID does not match returned mock report ID")
							}
							// Verify input status is failed.
							if status != model.StatusFailed {
								return errors.New("status is not set to FAILED after error")
							}
							return nil
						},
					},
				},
				notifier: &mockNotifier{
					mockFunc: func(subject, mssg string, fmt model.NotifFmt, recipients []string) error {
						// Return Err.
						return errMockNotify
					},
				},
				metricsClient: &mockMetricsClient{},
			},
			input: `
			{
				"team_info": {
					"id": "1",
					"name": "myTeam",
					"recipients": []
				},
				"data": {
					"scan_id": "1",
					"program_name": "progName"
				},
				"type": "scan",
				"auto_send": true
			}`,
			expectedMetricCalls: 1,
			expectedErr:         errMockNotify,
		},
		{
			name: "Should return ErrMockFinish",
			fields: fields{
				log: log,
				generateUCC: map[model.ReportType]GenerateUC{
					"scan": &mockGenerateUC{
						mockGenerateFunc: func(ctx context.Context, teamInfo teamInfo, reportData interface{}) (model.Report, error) {
							// All good.
							return &model.LiveReport{}, nil
						},
						mockFinishFunc: func(ctx context.Context, reportID, status string) error {
							// Return Err.
							return errMockFinish
						},
					},
				},
				notifier: &mockNotifier{
					mockFunc: func(subject, mssg string, fmt model.NotifFmt, recipients []string) error {
						// All good.
						return nil
					},
				},
				metricsClient: &mockMetricsClient{},
			},
			input: `
			{
				"team_info": {
					"id": "1",
					"name": "myTeam",
					"recipients": []
				},
				"data": {
					"scan_id": "1",
					"program_name": "progName"
				},
				"type": "scan",
				"auto_send": true
			}`,
			expectedMetricCalls: 2,
			expectedErr:         errMockFinish,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			processor, err := NewProcessor(tc.fields.log, tc.fields.generateUCC, tc.fields.notifier, tc.fields.metricsClient)
			if err != nil {
				t.Fatalf("Error building processor: %v", err)
			}

			err = processor.ProcessMessage(tc.input)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("Expected err: %v\nBut got: %v", tc.expectedErr, err)
			}
			metricCalls := tc.fields.metricsClient.(*mockMetricsClient).calls
			if metricCalls != tc.expectedMetricCalls {
				t.Fatalf("Expected metrics calls to be: %d\nBut got: %d", tc.expectedMetricCalls, metricCalls)
			}
		})
	}
}
