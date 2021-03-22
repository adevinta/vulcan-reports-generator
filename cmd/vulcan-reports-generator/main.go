/*
Copyright 2021 Adevinta
*/

package main

//go:generate rm -rf ../../vendor/github.com/adevinta/security-overview/resources/keep.go
//go:generate cp -pR ../../vendor/github.com/adevinta/security-overview/resources/. ../../_build/files/opt/vulcan-reports-generator/generators/scan/resources/.

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"sync"

	metrics "github.com/adevinta/vulcan-metrics-client"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/ses"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"github.com/adevinta/vulcan-reports-generator/pkg/api"
	"github.com/adevinta/vulcan-reports-generator/pkg/model"
	"github.com/adevinta/vulcan-reports-generator/pkg/notify"
	"github.com/adevinta/vulcan-reports-generator/pkg/queue"
	"github.com/adevinta/vulcan-reports-generator/pkg/report"
	"github.com/adevinta/vulcan-reports-generator/pkg/storage"
	"github.com/adevinta/vulcan-reports-generator/pkg/upload"
)

const (
	pg          = "postgres"
	pgConStrFmt = "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s"

	defRegion = "eu-west-1"
)

func main() {

	// Read config.
	cfgFilePath := flag.String("c", "./config.toml", "configuration file")
	flag.Parse()

	conf, err := parseConfig(*cfgFilePath)
	if err != nil {
		log.WithError(err).Fatal("Error reading configuration")
	}

	logger := setupLogger(*conf)

	// Build AWS session.
	awsSess := session.Must(session.NewSession())

	// Build uploader.
	if conf.S3.Region == "" {
		conf.S3.Region = defRegion
	}
	s3Config := &aws.Config{
		Region: aws.String(conf.S3.Region),
	}
	if conf.S3.Endpoint != "" {
		s3Config.WithEndpoint(conf.S3.Endpoint).WithS3ForcePathStyle(conf.S3.PathStyle)
	}
	uploader := upload.NewS3Uploader(s3manager.NewUploaderWithClient(s3.New(awsSess, s3Config)), logger)

	// Build notifier.
	if conf.SES.Region == "" {
		conf.SES.Region = defRegion
	}
	notifier, err := notify.NewSESNotifier(conf.SES, ses.New(awsSess, &aws.Config{
		Region: aws.String(conf.SES.Region),
	}))
	if err != nil {
		logger.WithError(err).Fatal("Error creating notifier")
	}
	// Build metrics client.
	metricsClient, err := metrics.NewClient()
	if err != nil {
		logger.WithError(err).Fatal("Error creating metrics client")
	}

	// Build DB.
	connStr := fmt.Sprintf(pgConStrFmt, conf.DB.Host, conf.DB.Port,
		conf.DB.User, conf.DB.Pass, conf.DB.Name, conf.DB.SSLMode)
	db, err := sql.Open(pg, connStr)
	if err != nil {
		logger.WithError(err).Fatal("Error connecting to DB")
	}
	defer db.Close()

	// Build generate Use Cases.
	generateUCC := map[model.ReportType]report.GenerateUC{}
	repositories := map[model.ReportType]storage.ReportsRepository{}

	for t, gconf := range conf.Generators {
		typ := model.ReportType(t)
		g, err := report.NewGenerator(t, gconf, logger, db)
		if err != nil {
			logger.WithError(err).WithFields(
				log.Fields{"type": t},
			).Fatal("Error building generator")
		}
		r, err := storage.NewReportsRepository(t, db)
		if err != nil {
			logger.WithError(err).WithFields(
				log.Fields{"type": t},
			).Fatal("Error building repository")
		}

		uc, err := report.NewGenerateUC(typ, logger, g, r)
		if err != nil {
			logger.WithError(err).WithFields(
				log.Fields{"type": t},
			).Fatal("Error building generate UC")
		}

		generateUCC[typ] = uc
		repositories[typ] = r
	}

	// Build processor.
	processor, err := report.NewProcessor(logger, generateUCC, uploader, notifier, metricsClient)
	if err != nil {
		logger.WithError(err).Fatal("Error creating queue processor")
	}

	sqsConsumerGroup, err := queue.NewSQSConsumerGroup(conf.SQS.NProcessors, conf.SQS.SQSConfig, processor, logger)
	if err != nil {
		logger.WithError(err).Fatal("Error creating queue consumer group")
	}

	// Build and start API.
	api := api.NewReportsAPI(api.NewReportsService(logger, notifier, repositories))
	go api.Start(conf.API.Port)

	// Start Consumer group.
	var wg sync.WaitGroup
	sqsConsumerGroup.Start(context.Background(), &wg)
	logger.Info("Started")
	wg.Wait()
}

func setupLogger(cfg config) *log.Logger {
	var logger = log.New()

	logger.SetFormatter(&log.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(parseLogLevel(cfg.Log.Level))
	logger.SetReportCaller(cfg.Log.AddCaller)

	return logger
}
