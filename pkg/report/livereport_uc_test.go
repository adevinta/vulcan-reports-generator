/*
Copyright 2021 Adevinta
*/

package report

import (
	"context"
	"errors"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/adevinta/vulcan-reports-generator/pkg/model"
	"github.com/adevinta/vulcan-reports-generator/pkg/storage"
)

var (
	errMockGet  = errors.New("ErrMockGet")
	errMockSave = errors.New("ErrMockSave")
)

// Generator mock.
type mockGenFunc func(ctx context.Context, teamInfo teamInfo, reportData interface{}) (interface{}, error)
type mockGenerator struct {
	Generator
	mockFunc mockGenFunc
}

func (g *mockGenerator) Generate(ctx context.Context, teamInfo teamInfo, reportData interface{}) (interface{}, error) {
	return g.mockFunc(ctx, teamInfo, reportData)
}

// Repository mock.
type mockGetFunc func(ctx context.Context, reportID string) (model.Report, error)
type mockSaveFunc func(ctx context.Context, report model.Report) error
type mockReportsRepository struct {
	storage.ReportsRepository
	mockGetFunc  mockGetFunc
	mockSaveFunc mockSaveFunc
}

func (r *mockReportsRepository) GetReport(ctx context.Context, reportID string) (model.Report, error) {
	return r.mockGetFunc(ctx, reportID)
}

func (r *mockReportsRepository) SaveReport(ctx context.Context, report model.Report) error {
	return r.mockSaveFunc(ctx, report)
}

type fields struct {
	generator  Generator
	repository storage.ReportsRepository
}

func TestGenerateUCLiveReport(t *testing.T) {
	testCases := []struct {
		name           string
		fields         fields
		teamInfo       teamInfo
		reportData     interface{}
		expectedReport model.Report
		expectedErr    error
	}{
		{
			name: "Happy path",
			fields: fields{
				repository: &mockReportsRepository{
					mockSaveFunc: func(ctx context.Context, report model.Report) error {
						r, ok := report.(*model.LiveReport)
						if !ok {
							return errors.New("Report is not LiveReport")
						}
						if r.TeamID != "11" || r.Status != model.StatusGenerating {
							return errors.New("Report does not match input")
						}
						// Set mock ID.
						r.ID = "12345"
						return nil
					},
				},
				generator: &mockGenerator{
					mockFunc: func(ctx context.Context, teamInfo teamInfo, reportData interface{}) (interface{}, error) {
						if teamInfo.ID != "1" || teamInfo.Name != "myTeam" {
							return nil, errors.New("TeamInfo data does not match input")
						}
						rd, ok := reportData.(liveReportRequest)
						if !ok || rd.TeamID != "11" || rd.DateFrom != "2020-09-01" {
							return nil, errors.New("reportData does not match input")
						}
						return liveReportData{
							EmailSubject: "emailSubject",
							EmailBody:    "emailBody",
						}, nil
					},
				},
			},
			teamInfo: teamInfo{
				ID:   "1",
				Name: "myTeam",
			},
			reportData: liveReportRequest{
				TeamID:   "11",
				DateFrom: "2020-09-01",
				DateTo:   "2020-09-07",
			},
			expectedReport: &model.LiveReport{
				BaseReport: model.BaseReport{
					ID:     "12345",
					Status: model.StatusGenerating,
					Notification: model.Notification{
						Subject: "emailSubject",
						Body:    "emailBody",
						Fmt:     model.NotifFmtHTML,
					},
				},
				TeamID:   "11",
				DateFrom: "2020-09-01",
				DateTo:   "2020-09-07",
			},
		},
		{
			name: "Should return ErrInvalidRequest",
			fields: fields{
				repository: &mockReportsRepository{},
				generator:  &mockGenerator{},
			},
			reportData:  struct{}{},
			expectedErr: ErrInvalidRequest,
		},
		{
			name: "Should return ErrInvalidRequest, void TeamID",
			fields: fields{
				repository: &mockReportsRepository{},
				generator:  &mockGenerator{},
			},
			reportData: liveReportRequest{
				DateFrom: "2020-09-01",
			},
			expectedErr: ErrInvalidRequest,
		},
		{
			name: "Should return ErrInvalidRequest, void DateFrom",
			fields: fields{
				repository: &mockReportsRepository{},
				generator:  &mockGenerator{},
			},
			reportData: liveReportRequest{
				TeamID: "11",
			},
			expectedErr: ErrInvalidRequest,
		},
		{
			name: "Should return ErrMockSave",
			fields: fields{
				repository: &mockReportsRepository{
					mockSaveFunc: func(ctx context.Context, report model.Report) error {
						// Return Err.
						return errMockSave
					},
				},
				generator: &mockGenerator{},
			},
			reportData: liveReportRequest{
				TeamID:   "11",
				DateFrom: "2020-09-01",
				DateTo:   "2020-09-07",
			},
			expectedErr: errMockSave,
		},
		{
			name: "Should return ErrMockGen",
			fields: fields{
				repository: &mockReportsRepository{
					mockSaveFunc: func(ctx context.Context, report model.Report) error {
						// All good.
						return nil
					},
					mockGetFunc: func(ctx context.Context, reportID string) (model.Report, error) {
						// All good.
						return &model.LiveReport{}, nil
					},
				},
				generator: &mockGenerator{
					mockFunc: func(ctx context.Context, teamInfo teamInfo, reportData interface{}) (interface{}, error) {
						// Return Err.
						return nil, errMockGen
					},
				},
			},
			reportData: liveReportRequest{
				TeamID:   "11",
				DateFrom: "2020-09-01",
				DateTo:   "2020-09-07",
			},
			expectedErr: errMockGen,
		},
	}

	ctx := context.Background()
	logger := log.New()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			genUC := &livereportUC{
				log:        logger,
				generator:  tc.fields.generator,
				repository: tc.fields.repository,
			}

			report, err := genUC.Generate(ctx, tc.teamInfo, tc.reportData)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("Expected err: %v\nBut got: %v", tc.expectedErr, err)
			}
			if !reflect.DeepEqual(report, tc.expectedReport) {
				t.Fatalf("Expected report: %v\nBut got: %v", tc.expectedReport, report)
			}
		})
	}
}

func TestFinishLiveReport(t *testing.T) {
	testCases := []struct {
		name        string
		fields      fields
		reportID    string
		status      string
		expectedErr error
	}{
		{
			name: "Happy path",
			fields: fields{
				repository: &mockReportsRepository{
					mockGetFunc: func(ctx context.Context, reportID string) (model.Report, error) {
						return &model.LiveReport{
							BaseReport: model.BaseReport{
								ID:          reportID,
								Status:      model.StatusGenerating,
								DeliveredTo: []string{},
							},
							TeamID: "111",
						}, nil
					},
					mockSaveFunc: func(ctx context.Context, report model.Report) error {
						r, ok := report.(*model.LiveReport)
						if !ok {
							return errors.New("Report is not LiveReport")
						}
						if r.ID != "1" || r.TeamID != "111" {
							return errors.New("Report data does not match input")
						}
						if r.Status != model.StatusFinished {
							return errors.New("Status or recipients not updated")
						}
						return nil
					},
				},
			},
			reportID: "1",
			status:   model.StatusFinished,
		},
		{
			name: "Should return ErrMockGet",
			fields: fields{
				repository: &mockReportsRepository{
					mockGetFunc: func(ctx context.Context, reportID string) (model.Report, error) {
						// Return Err.
						return nil, errMockGet
					},
				},
			},
			expectedErr: errMockGet,
		},
		{
			name: "Should return ErrMockSave",
			fields: fields{
				repository: &mockReportsRepository{
					mockGetFunc: func(ctx context.Context, reportID string) (model.Report, error) {
						// All good.
						return &model.LiveReport{}, nil
					},
					mockSaveFunc: func(ctx context.Context, report model.Report) error {
						// Return Err.
						return errMockSave
					},
				},
			},
			expectedErr: errMockSave,
		},
	}

	ctx := context.Background()
	logger := log.New()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			genUC := &livereportUC{
				log:        logger,
				generator:  tc.fields.generator,
				repository: tc.fields.repository,
			}

			err := genUC.Finish(ctx, tc.reportID, tc.status)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("Expected err: %v\nBut got: %v", tc.expectedErr, err)
			}
		})
	}
}
