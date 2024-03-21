/*
Copyright 2021 Adevinta
*/

package report

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	liveEmailSubjectFmt = "%s - %s"
)

type liveReportGeneratorCfg struct {
	EmailSubject      string `toml:"email_subject" mapstructure:"email_subject"`
	EmailTemplateFile string `toml:"email_template_file" mapstructure:"email_template_file"`
}

// liveReportRequest is the expected
// data supplied in the generation
// request for a live report.
type liveReportRequest struct {
	TeamID        string `mapstructure:"team_id"`
	Info          int    `mapstructure:"info"`
	Low           int    `mapstructure:"low"`
	Medium        int    `mapstructure:"medium"`
	High          int    `mapstructure:"high"`
	Critical      int    `mapstructure:"critical"`
	InfoDiff      int    `mapstructure:"info_diff"`
	LowDiff       int    `mapstructure:"low_diff"`
	MediumDiff    int    `mapstructure:"medium_diff"`
	HighDiff      int    `mapstructure:"high_diff"`
	CriticalDiff  int    `mapstructure:"critical_diff"`
	InfoFixed     int    `mapstructure:"info_fixed"`
	LowFixed      int    `mapstructure:"low_fixed"`
	MediumFixed   int    `mapstructure:"medium_fixed"`
	HighFixed     int    `mapstructure:"high_fixed"`
	CriticalFixed int    `mapstructure:"critical_fixed"`
	DateFrom      string `mapstructure:"date_from"`
	DateTo        string `mapstructure:"date_to"`
	URL           string `mapstructure:"live_report_url"`
}

type liveReportData struct {
	EmailSubject string
	EmailBody    string
}

type liveReportGenerator struct {
	cfg      liveReportGeneratorCfg
	template *template.Template
	log      *log.Logger
}

// newLiveReportGenerator creates a new Generator for Live reports.
func newLiveReportGenerator(cfg liveReportGeneratorCfg, log *log.Logger) (Generator, error) {
	g := &liveReportGenerator{
		cfg: cfg,
		log: log,
	}

	// Verify template on init.
	tmplPath, err := g.getTeamplePath()
	if err != nil {
		return nil, err
	}
	tmpl, err := os.ReadFile(tmplPath)
	if err != nil {
		return nil, err
	}
	g.template = template.Must(template.New("Email").Funcs(template.FuncMap{
		"trendArrow":    trendArrow,
		"trendColor":    trendColor,
		"severityColor": severityColor,
	}).Parse(string(tmpl)))

	return g, nil
}

// Generate ...
func (g *liveReportGenerator) Generate(ctx context.Context, teamInfo teamInfo, reportData interface{}) (interface{}, error) {
	liveReportReq, err := parseLiveReportReq(reportData)
	if err != nil {
		return nil, err
	}

	// Print report files.
	g.log.WithFields(log.Fields{
		"teamID": teamInfo.ID,
		"type":   "liveReport",
	}).Info("Generating report")

	liveReportData, err := g.Print(teamInfo, liveReportReq)
	if err != nil {
		return nil, err
	}

	liveReportData.EmailSubject = fmt.Sprintf(liveEmailSubjectFmt,
		g.cfg.EmailSubject, teamInfo.Name)
	return liveReportData, nil
}

func (g *liveReportGenerator) Print(teamInfo teamInfo, liveReportReq liveReportRequest) (liveReportData, error) {
	type Severity struct {
		Description   string
		TotalFindings int
		NewFindings   int
		FixedFindings int
	}
	type Report struct {
		TeamName         string
		StartDate        string
		EndDate          string
		LinkToLiveReport string
		Severities       []Severity
		ImportantFixed   int
	}

	r := Report{
		TeamName:         teamInfo.Name,
		StartDate:        liveReportReq.DateFrom,
		EndDate:          liveReportReq.DateTo,
		LinkToLiveReport: liveReportReq.URL,
		Severities: []Severity{
			{Description: "Critical", TotalFindings: liveReportReq.Critical, NewFindings: liveReportReq.CriticalDiff, FixedFindings: liveReportReq.CriticalFixed},
			{Description: "High", TotalFindings: liveReportReq.High, NewFindings: liveReportReq.HighDiff, FixedFindings: liveReportReq.HighFixed},
			{Description: "Medium", TotalFindings: liveReportReq.Medium, NewFindings: liveReportReq.MediumDiff, FixedFindings: liveReportReq.MediumFixed},
			{Description: "Low", TotalFindings: liveReportReq.Low, NewFindings: liveReportReq.LowDiff, FixedFindings: liveReportReq.LowFixed},
		},
		ImportantFixed: liveReportReq.CriticalFixed + liveReportReq.HighFixed,
	}

	var output []byte
	buf := bytes.NewBuffer(output)
	err := g.template.Execute(buf, r)
	if err != nil {
		return liveReportData{}, err
	}
	emailContent := string(buf.Bytes())

	return liveReportData{
		EmailBody: emailContent,
	}, nil
}

func (g *liveReportGenerator) getTeamplePath() (string, error) {
	// If configured out dir
	// is relative path, get wd.
	var wd string
	var err error
	if !strings.HasPrefix(g.cfg.EmailTemplateFile, string(os.PathSeparator)) {
		wd, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	return filepath.Join(wd, g.cfg.EmailTemplateFile), nil
}

func parseLiveReportReq(liveReportData interface{}) (liveReportRequest, error) {
	liveReportReq, ok := liveReportData.(liveReportRequest)
	if !ok || liveReportReq.TeamID == "" || liveReportReq.DateFrom == "" || liveReportReq.DateTo == "" {
		return liveReportRequest{}, ErrInvalidRequest
	}
	return liveReportReq, nil
}

func trendArrow(new int, fixed int) string {
	if new > fixed {
		return "▲"
	} else if fixed > new {
		return "▼"
	} else {
		return "="
	}
}

func trendColor(new int, fixed int) string {
	if new > fixed {
		return "red"
	} else if fixed > new {
		return "green"
	} else {
		return "grey"
	}
}

func severityColor(severity string) string {
	switch severity {
	case "Critical":
		return "purple"
	case "High":
		return "red"
	case "Medium":
		return "orange"
	case "Low":
		return "yellow"
	default:
		return "black"
	}
}
