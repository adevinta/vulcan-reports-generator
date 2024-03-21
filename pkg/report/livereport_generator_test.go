/*
Copyright 2021 Adevinta
*/

package report

import (
	"context"
	"os"
	"reflect"
	"testing"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

func TestGenerateLiveReport(t *testing.T) {
	os.WriteFile("/tmp/template", []byte("notif"), 0x755)

	mockCfg := liveReportGeneratorCfg{
		EmailSubject:      "[UnitTest] Live Report",
		EmailTemplateFile: "/tmp/template",
	}

	mockLog := log.New()
	type fields struct {
		cfg          liveReportGeneratorCfg
		retEmailBody string
		retErr       error
	}
	type input struct {
		ctx        context.Context
		reportID   string
		teamInfo   teamInfo
		reportData interface{}
	}

	testCases := []struct {
		name                   string
		fields                 fields
		input                  input
		expectedLiveReportData interface{}
		expectedErr            error
	}{
		{
			name: "Happy path",
			fields: fields{
				cfg:          mockCfg,
				retEmailBody: "notif",
			},
			input: input{
				reportData: liveReportRequest{
					TeamID:   "1",
					DateFrom: "2020-09-01",
					DateTo:   "2020-09-07",
				},
				teamInfo: teamInfo{
					Name: "TeamName",
				},
			},
			expectedLiveReportData: liveReportData{
				EmailSubject: "[UnitTest] Live Report - TeamName",
				EmailBody:    "notif",
			},
		},
		{
			name: "Should return ErrInvalidRequest, bad req fmt",
			fields: fields{
				cfg: mockCfg,
			},
			input: input{
				reportData: liveReportRequest{
					DateFrom: "2020-09-01",
				},
			},
			expectedErr: ErrInvalidRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			generator, _ := newLiveReportGenerator(tc.fields.cfg, mockLog)

			reportData, err := generator.Generate(tc.input.ctx, tc.input.teamInfo, tc.input.reportData)
			if err != nil {
				if tc.expectedErr == nil || tc.expectedErr != err {
					t.Fatalf("Expected error: %v\nBut got: %v", tc.expectedErr, err)
				}
			}
			if !reflect.DeepEqual(reportData, tc.expectedLiveReportData) {
				t.Fatalf("Expected live report data: %v\nBut got: %v", tc.expectedLiveReportData, reportData)
			}
		})
	}
}
