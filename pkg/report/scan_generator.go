/*
Copyright 2021 Adevinta
*/

package report

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/adevinta/vulcan-reports-generator/pkg/model"
	fileutils "github.com/adevinta/vulcan-reports-generator/pkg/utils/files"
)

const (
	emailSubjectFmt = "[%s] %s - %s - %s"
)

type scanGeneratorCfg struct {
	PrinterCfgFile string `toml:"printer_cfg_file" mapstructure:"printer_cfg_file"`
	EmailSubject   string `toml:"email_subject" mapstructure:"email_subject"`
	OutputDir      string `toml:"output_dir" mapstructure:"output_dir"`
	PublicBucket   string `toml:"public_bucket" mapstructure:"public_bucket"`
	PrivateBucket  string `toml:"private_bucket" mapstructure:"private_bucket"`
}

// scanRequest is the expected
// data supplied in the generation
// request for a scan report.
type scanRequest struct {
	ScanID      string `mapstructure:"scan_id"`
	ProgramName string `mapstructure:"program_name"`
}

type scanReportData struct {
	ReportURL     string
	ReportJSONURL string
	Files         []model.FileInfo
	EmailSubject  string
	EmailBody     string
	Action        string
	Risk          int
}

type scanGenerator struct {
	cfg     scanGeneratorCfg
	log     *log.Logger
	printer scanPrinter
}

// newScanGenerator creates a new Generator for Scan reports.
func newScanGenerator(cfg scanGeneratorCfg, log *log.Logger, printer scanPrinter) Generator {
	return &scanGenerator{
		cfg:     cfg,
		log:     log,
		printer: printer,
	}
}

// Generate generates the overview and full report files for the given team info and report data.
// Report files are generated in the configured output dir for the generator.
// Generation is performed by using security-overview library.
func (g *scanGenerator) Generate(ctx context.Context, teamInfo teamInfo, reportData interface{}) (interface{}, error) {
	scanReq, err := parseScanReq(reportData)
	if err != nil {
		return nil, err
	}

	// Print report files.
	g.log.WithFields(log.Fields{
		"teamID": teamInfo.ID,
		"type":   "scan",
		"scanID": scanReq.ScanID,
	}).Info("Generating report")
	scanReportData, err := g.printer.Print(teamInfo, scanReq)
	if err != nil {
		return nil, err
	}

	// Retrieve report files
	// and return report data.
	g.log.WithFields(log.Fields{
		"teamID": teamInfo.ID,
		"type":   "scan",
		"scanID": scanReq.ScanID,
	}).Debug("Retrieving report files")
	reportFiles, err := g.getReportFiles(scanReq.ScanID)
	if err != nil {
		return nil, err
	}

	scanReportData.Files = reportFiles
	scanReportData.EmailSubject = fmt.Sprintf(emailSubjectFmt,
		scanReportData.Action, g.cfg.EmailSubject, teamInfo.Name, scanReq.ProgramName)
	return scanReportData, nil
}

// getReportFiles returns a list of FileInfo objects for the generated report
// files (public and private) for the specified scanID.
func (g *scanGenerator) getReportFiles(scanID string) ([]model.FileInfo, error) {
	reportDir, err := g.getReportDir(scanID)
	if err != nil {
		return nil, err
	}

	var files []model.FileInfo

	// Public contents.
	// output/scanID/<public-bucket>/* -> to public bucket
	publicDir := filepath.Join(reportDir, g.cfg.PublicBucket)
	publicFiles, err := buildFInfoFiles(publicDir, g.cfg.PublicBucket)
	if err != nil {
		return nil, err
	}
	g.log.WithFields(log.Fields{
		"type":   "scan",
		"scanID": scanID,
		"files":  publicFiles,
	}).Trace("Retrieved report public files")
	files = append(files, publicFiles...)

	// Private contents.
	// output/scanID/<private-bucket>/* -> to private bucket
	privateDir := filepath.Join(reportDir, g.cfg.PrivateBucket)
	privateFiles, err := buildFInfoFiles(privateDir, g.cfg.PrivateBucket)
	if err != nil {
		return nil, err
	}
	g.log.WithFields(log.Fields{
		"type":   "scan",
		"scanID": scanID,
		"files":  privateFiles,
	}).Trace("Retrieved report private files")
	files = append(files, privateFiles...)

	return files, nil
}

// buildFInfoFiles iterates recursively through the specified dir
// and returns a slice of FileInfo objects for each one of them,
// setting the destBucket to the input one, and the destPath (key)
// to the full file path minus the input dir path.
// E.g.:
//    dir = /home/user/reportDir
//    f   = /home/user/reportDir/subDir/file.ext
// => destPath = subdir/file.ext
func buildFInfoFiles(dir, destBucket string) ([]model.FileInfo, error) {
	var files []model.FileInfo

	filePaths, err := fileutils.ListDirFiles(dir)
	if err != nil {
		return nil, err
	}
	for _, f := range filePaths {
		// fRelPath is the relative path
		// from the base dir path.
		fRelPath, err := filepath.Rel(dir, f)
		if err != nil {
			return nil, err
		}
		fi := model.FileInfo{
			FilePath:   f,
			DestBucket: destBucket,
			DestPath:   fRelPath,
		}
		files = append(files, fi)
	}

	return files, nil
}

func (g *scanGenerator) getReportDir(scanID string) (string, error) {
	// If configured out dir
	// is relative path, get wd.
	var wd string
	var err error
	if !strings.HasPrefix(g.cfg.OutputDir, string(os.PathSeparator)) {
		wd, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	return filepath.Join(wd, g.cfg.OutputDir, scanID), nil
}

func parseScanReq(scanData interface{}) (scanRequest, error) {
	scanReq, ok := scanData.(scanRequest)
	if !ok || scanReq.ProgramName == "" || scanReq.ScanID == "" {
		return scanRequest{}, ErrInvalidRequest
	}
	return scanReq, nil
}
