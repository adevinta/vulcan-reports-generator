/*
Copyright 2021 Adevinta
*/

package report

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"github.com/adevinta/vulcan-reports-generator/pkg/model"
)

type mockScanReportPrinterCfg struct {
	outputDir  string
	pubBucket  string
	privBucket string
}

type mockScanReportPrinter struct {
	cfg mockScanReportPrinterCfg

	pubFiles  []string
	privFiles []string

	retReportURL     string
	retReportURLJSON string
	retEmailBody     string
	retAction        string
	retRisk          int

	retErr error
}

func (m *mockScanReportPrinter) Print(teamInfo teamInfo, scanReq scanRequest) (scanReportData, error) {
	if m.retErr != nil {
		return scanReportData{}, m.retErr
	}

	// Mock report files generation.
	reportData := scanReportData{}

	// Create scan report pub dir.
	// Create one mock file per each pubFiles defined.
	// Output follows pattern /{out_dir}/{scan_id}/{pub_bucket}/{pubFile}
	pubDir := filepath.Join(m.cfg.outputDir, scanReq.ScanID, m.cfg.pubBucket)
	err := os.MkdirAll(pubDir, os.ModePerm)
	if err != nil {
		return scanReportData{}, err
	}
	for _, pubF := range m.pubFiles {
		pubFPath := filepath.Join(pubDir, pubF)
		os.Create(pubFPath)
		reportData.Files = append(reportData.Files, model.FileInfo{
			FilePath: pubFPath,
		})
	}

	// Create scan report priv dir.
	// Create one mock file per each privFiles defined.
	// Output follows pattern /{out_dir}/{scan_id}/{priv_bucket}/{privFile}
	privDir := filepath.Join(m.cfg.outputDir, scanReq.ScanID, m.cfg.privBucket)
	err = os.MkdirAll(privDir, os.ModePerm)
	if err != nil {
		return scanReportData{}, err
	}
	for _, privF := range m.privFiles {
		privFPath := filepath.Join(privDir, privF)
		os.Create(privFPath)
		reportData.Files = append(reportData.Files, model.FileInfo{
			FilePath: privFPath,
		})
	}

	// Set rest of report fields
	// to mock's defined ones.
	reportData.Risk = m.retRisk
	reportData.Action = m.retAction
	reportData.EmailBody = m.retEmailBody
	reportData.ReportURL = m.retReportURL
	reportData.ReportJSONURL = m.retReportURLJSON

	return reportData, nil
}

func TestGenerate(t *testing.T) {
	// Create tmp test dir for mock printer report output.
	testDir := fmt.Sprintf("/tmp/test-%s", time.Now().String())
	err := os.Mkdir(testDir, os.ModePerm)
	if err != nil {
		t.Fatalf("Error setting up test dir: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Constants for mock
	// output files generation.
	pubBucket := "pub"
	privBucket := "priv"
	// Set scan generator config to
	// match test dir output.
	mockCfg := scanGeneratorCfg{
		EmailSubject:  "[UnitTest] SecOverview",
		OutputDir:     testDir,
		PublicBucket:  pubBucket,
		PrivateBucket: privBucket,
	}
	// Set printer config to match
	// scan generator config output.
	mockPrinterCfg := mockScanReportPrinterCfg{
		outputDir:  mockCfg.OutputDir,
		pubBucket:  mockCfg.PublicBucket,
		privBucket: mockCfg.PrivateBucket,
	}

	mockLog := log.New()
	mockPrintErr := errors.New("ErrPrint")

	type fields struct {
		cfg     scanGeneratorCfg
		printer scanPrinter
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
		expectedScanReportData interface{}
		expectedErr            error
	}{
		{
			name: "Happy path",
			fields: fields{
				cfg: mockCfg,
				printer: &mockScanReportPrinter{
					cfg:              mockPrinterCfg,
					pubFiles:         []string{"pubF1"},
					privFiles:        []string{"privF1"},
					retReportURL:     "reportURL",
					retReportURLJSON: "reportJSONRUL",
					retEmailBody:     "notif",
					retAction:        "ACTION REQUIRED",
					retRisk:          3,
				},
			},
			input: input{
				reportData: scanRequest{
					ScanID:      "1",
					ProgramName: "ProgName",
				},
				teamInfo: teamInfo{
					Name: "TeamName",
				},
			},
			expectedScanReportData: scanReportData{
				Files: []model.FileInfo{
					{
						FilePath:   filepath.Join(mockCfg.OutputDir, "1", mockCfg.PublicBucket, "pubF1"),
						DestBucket: mockCfg.PublicBucket,
						DestPath:   "pubF1",
					},
					{
						FilePath:   filepath.Join(mockCfg.OutputDir, "1", mockCfg.PrivateBucket, "privF1"),
						DestBucket: mockCfg.PrivateBucket,
						DestPath:   "privF1",
					},
				},
				EmailSubject:  "[ACTION REQUIRED] [UnitTest] SecOverview - TeamName - ProgName",
				EmailBody:     "notif",
				ReportURL:     "reportURL",
				ReportJSONURL: "reportJSONRUL",
				Action:        "ACTION REQUIRED",
				Risk:          3,
			},
		},
		{
			name: "Should return ErrInvalidRequest, bad req fmt",
			fields: fields{
				cfg: mockCfg,
			},
			input: input{
				reportData: scanRequest{
					ProgramName: "program2",
				},
			},
			expectedErr: ErrInvalidRequest,
		},
		{
			name: "Should return print error",
			fields: fields{
				cfg: mockCfg,
				printer: &mockScanReportPrinter{
					retErr: mockPrintErr,
				},
			},
			input: input{
				reportData: scanRequest{
					ScanID:      "44",
					ProgramName: "mockProgram",
				},
			},
			expectedErr: mockPrintErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			generator := newScanGenerator(tc.fields.cfg, mockLog, tc.fields.printer)

			reportData, err := generator.Generate(tc.input.ctx, tc.input.teamInfo, tc.input.reportData)
			if err != nil {
				if tc.expectedErr == nil || tc.expectedErr != err {
					t.Fatalf("Expected error: %v\nBut got: %v", tc.expectedErr, err)
				}
			}
			if !reflect.DeepEqual(reportData, tc.expectedScanReportData) {
				t.Fatalf("Expected scan report data: %v\nBut got: %v", tc.expectedScanReportData, reportData)
			}
		})
	}
}
