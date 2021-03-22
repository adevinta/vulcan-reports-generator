/*
Copyright 2021 Adevinta
*/

package queue

import (
	"context"
	"encoding/json"
	"errors"
	"runtime/debug"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	log "github.com/sirupsen/logrus"
)

const (
	maxNumberOfMsg = 10
	defSQSWaitTime = 0
)

// SQSConfig is the configuration required for an SQSConsumer.
type SQSConfig struct {
	QueueArn    string `toml:"queue_arn"`
	Timeout     int64  `toml:"timeout"`
	MaxWaitTime int64  `toml:"wait_time"`
	QueueName   string `toml:"queue_name"`
	Endpoint    string `toml:"endpoint"`
}

// SQSConsumer is the SQS implementation of the QueueConsumer interface.
type SQSConsumer struct {
	config      SQSConfig
	sqsURL      string
	sqsWaitTime int64
	sqs         sqsiface.SQSAPI
	processor   Processor
	logger      *log.Logger
}

// SQSConsumerGroup is a group of SQSConsumers.
type SQSConsumerGroup struct {
	consumers []*SQSConsumer
}

// NewSQSConsumerGroup creates a new SQSConsumerGroup.
func NewSQSConsumerGroup(nConsumers uint8, config SQSConfig, processor Processor, logger *log.Logger) (*SQSConsumerGroup, error) {
	var consumerGroup SQSConsumerGroup

	awsSess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	arn, err := arn.Parse(config.QueueArn)
	if err != nil {
		return nil, err
	}

	awsCfg := aws.NewConfig()

	if arn.Region != "" {
		awsCfg = awsCfg.WithRegion(arn.Region)
	}
	if config.Endpoint != "" {
		awsCfg = awsCfg.WithEndpoint(config.Endpoint)
	}

	input := sqs.GetQueueUrlInput{
		QueueName: aws.String(arn.Resource),
	}
	if arn.AccountID != "" {
		input.SetQueueOwnerAWSAccountId(arn.AccountID)
	}

	sqsSvc := sqs.New(awsSess, awsCfg)
	sqsURLData, err := sqsSvc.GetQueueUrl(&input)
	if err != nil {
		return nil, err
	}

	var consumers []*SQSConsumer
	for i := uint8(0); i < nConsumers; i++ {
		consumers = append(consumers, &SQSConsumer{
			config:      config,
			sqsURL:      *sqsURLData.QueueUrl,
			sqsWaitTime: defSQSWaitTime,
			sqs:         sqsSvc,
			processor:   processor,
			logger:      logger,
		})
	}

	consumerGroup.consumers = consumers
	return &consumerGroup, nil
}

// Start makes the consumer group start reading and processing messages from the queue.
func (g *SQSConsumerGroup) Start(ctx context.Context, wg *sync.WaitGroup) {
	for _, consumer := range g.consumers {
		wg.Add(1)
		go consumer.start(ctx, wg)
	}
}

func (c *SQSConsumer) start(ctx context.Context, wg *sync.WaitGroup) {
	defer func() {
		if err := recover(); err != nil {
			c.logger.WithFields(log.Fields{
				"err":   err,
				"trace": string(debug.Stack()),
			}).Error("Consumer stopping due to panic err")
		}

		wg.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := c.readAndProcess(ctx); err != nil {
				c.logger.WithError(err).Error("Error reading SQS messages")
			}
		}
	}
}

func (c *SQSConsumer) readAndProcess(ctx context.Context) error {
	mssgs, err := c.readMssgs(ctx)
	if err != nil {
		return err
	}

	// Adjust SQS wait time based on
	// number of retrieved messages
	if len(mssgs) == 0 {
		c.sqsWaitTime = c.config.MaxWaitTime
	} else {
		c.sqsWaitTime = defSQSWaitTime
	}

	for _, mssg := range mssgs {
		// Check for invalid mssg
		mssgBody, err := validateMssg(mssg)
		if err != nil {
			c.logger.WithError(err).WithFields(log.Fields{
				"mssg": mssg,
			}).Error("Invalid SQS message")

			if err = c.deleteMessage(mssg); err != nil {
				c.logger.WithError(err).Error("Error deleting processed message")
			}
			continue
		}

		// If message is valid, process it
		if err = c.processor.ProcessMessage(mssgBody); err != nil {
			c.logger.WithError(err).WithFields(log.Fields{
				"body":  mssgBody,
				"attrs": mssg.Attributes,
			}).Error("Error processing SQS message")
			continue
		}

		// Delete it
		if err = c.deleteMessage(mssg); err != nil {
			c.logger.WithError(err).Error("Error deleting processed message")
		}
	}

	return nil
}

func (c *SQSConsumer) readMssgs(ctx context.Context) ([]*sqs.Message, error) {
	receiveQuery := sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(c.sqsURL),
		MaxNumberOfMessages: aws.Int64(maxNumberOfMsg),
		WaitTimeSeconds:     aws.Int64(c.sqsWaitTime),
		VisibilityTimeout:   aws.Int64(c.config.Timeout),
		AttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
	}
	mssgsResp, err := c.sqs.ReceiveMessageWithContext(ctx, &receiveQuery)
	if err != nil {
		return nil, err
	}

	return mssgsResp.Messages, nil
}

func (c *SQSConsumer) deleteMessage(mssg *sqs.Message) error {
	_, err := c.sqs.DeleteMessage(&sqs.DeleteMessageInput{
		ReceiptHandle: mssg.ReceiptHandle,
		QueueUrl:      aws.String(c.sqsURL),
	})

	return err
}

func validateMssg(mssg *sqs.Message) (string, error) {
	if mssg == nil || mssg.Body == nil {
		return "", errors.New("unpexpected nil message")
	}
	return removeSNSEnvelope(*mssg.Body)
}

// removeSNSEnvelope removes the SNS envelope around the
// actual body message because messages received in queue
// are pushed through an SNS topic.
func removeSNSEnvelope(snsNotif string) (string, error) {
	var mold = struct {
		Message string `json:"Message"`
	}{}
	if err := json.Unmarshal([]byte(snsNotif), &mold); err != nil {
		return "", errors.New("unexpected mssg format: Expected SNS envelope")
	}
	return mold.Message, nil
}
