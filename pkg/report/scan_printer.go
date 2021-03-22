/*
Copyright 2021 Adevinta
*/

package report

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	secoverview "github.com/adevinta/security-overview"
	secoverviewreport "github.com/adevinta/security-overview/report"
)

const (
	htmlExt = ".html"
	jsonExt = ".json"
)

// scanPrinter is an interface to abstract
// the security overview implementation of report files printing.
type scanPrinter interface {
	Print(teamInfo teamInfo, scanReq scanRequest) (scanReportData, error)
}

type secOverviewPrinter struct {
	cfgFile string
}

func newScanPrinter(cfgFile string) scanPrinter {
	return &secOverviewPrinter{
		cfgFile: cfgFile,
	}
}

func (p *secOverviewPrinter) Print(teamInfo teamInfo, scanReq scanRequest) (scanReportData, error) {
	scanReport, err := secoverview.NewDetailedReport(p.cfgFile, teamInfo.Name, scanReq.ScanID, teamInfo.ID)
	if err != nil {
		return scanReportData{}, err
	}

	err = scanReport.GenerateLocalFiles()
	if err != nil {
		return scanReportData{}, err
	}
	defer cleanDbugFile(teamInfo.Name)

	emailBodyBuf, err := ioutil.ReadFile(scanReport.Email)
	if err != nil {
		return scanReportData{}, err
	}

	return scanReportData{
		ReportURL:     scanReport.URL,
		ReportJSONURL: parseReportURLToJSON(scanReport.URL),
		EmailBody:     string(emailBodyBuf),
		Action:        secoverviewreport.RiskToActionString(scanReport.Risk),
		Risk:          scanReport.Risk,
	}, nil
}

// security overview generates a file  with name
// <teamName>.json in it's execution directory and
// does not remove it, so we have to clean it from here.
func cleanDbugFile(teamName string) {
	os.Remove(fmt.Sprintf("%s%s", teamName, jsonExt))
}

func parseReportURLToJSON(url string) string {
	if strings.HasSuffix(strings.ToLower(url), htmlExt) {
		return url[:len(url)-5] + jsonExt
	}
	return ""
}
