/*
Copyright 2021 Adevinta
*/

package queue

import (
	"context"
	"errors"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

const (
	sqsMaxWaitTime = 20

	mockSNSMssg = `
	{
		"Type" : "SubscriptionConfirmation",
		"MessageId" : "165545c9-2a5c-472c-8df2-7ff2be2b3b1b",
		"Token" : "2336412f37...",
		"TopicArn" : "arn:aws:sns:us-west-2:123456789012:MyTopic",
		"Message" : "You have chosen to subscribe to the topic...",
		"SubscribeURL" : "https://sns.us-west-2.amazonaws.com/?Action=...",
		"Timestamp" : "2012-04-26T20:45:04.751Z",
		"SignatureVersion" : "1",
		"Signature" : "EXAMPLEpH+DcEwjAPg8O9mY8dReBSwksfg2S7WKQcikcNK=...",
		"SigningCertURL" : "https://sns.us-west-2.amazonaws.com/mock.pem"
	}`
)

type sqsMock struct {
	sqsiface.SQSAPI

	returnMssgs    uint8
	wantReceiveErr bool
	wantDeleteErr  bool
	receiveCalls   uint8
	deleteCalls    uint8
}

func (m *sqsMock) ReceiveMessageWithContext(aws.Context, *sqs.ReceiveMessageInput, ...request.Option) (*sqs.ReceiveMessageOutput, error) {
	m.receiveCalls++

	if m.wantReceiveErr {
		return nil, errors.New("mockErr")
	}

	resp := &sqs.ReceiveMessageOutput{}
	mssgs := make([]*sqs.Message, 0)

	for i := uint8(0); i < m.returnMssgs; i++ {
		mssgs = append(mssgs, &sqs.Message{
			Body: aws.String(mockSNSMssg),
		})
	}
	resp.Messages = mssgs

	return resp, nil
}

func (m *sqsMock) DeleteMessage(*sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	m.deleteCalls++

	if m.wantDeleteErr {
		return nil, errors.New("mockErr")
	}

	return nil, nil
}

type mockProcessor struct {
	processCalls uint8
}

func (p *mockProcessor) ProcessMessage(mssg string) error {
	p.processCalls++
	return nil
}

func TestReadAndProcess(t *testing.T) {
	type fields struct {
		config    SQSConfig
		sqs       *sqsMock
		processor *mockProcessor
		logger    *log.Logger
	}

	logger := log.New()

	sqsConf := SQSConfig{
		MaxWaitTime: sqsMaxWaitTime,
	}

	tests := []struct {
		name              string
		fields            fields
		expectedErr       bool
		expectedDelCalls  uint8
		expectedRecCalls  uint8
		expectedProcCalls uint8
		expectedWaitTime  int64
	}{
		{
			name: "Should return err reading messages",
			fields: fields{
				config: sqsConf,
				sqs: &sqsMock{
					wantReceiveErr: true,
				},
				processor: &mockProcessor{},
				logger:    logger,
			},
			expectedErr:       true,
			expectedDelCalls:  0,
			expectedRecCalls:  1,
			expectedProcCalls: 0,
			expectedWaitTime:  0,
		},
		{
			name: "Should set SQS wait time to max",
			fields: fields{
				config: sqsConf,
				sqs: &sqsMock{
					returnMssgs: 0,
				},
				processor: &mockProcessor{},
				logger:    logger,
			},
			expectedErr:       false,
			expectedDelCalls:  0,
			expectedRecCalls:  1,
			expectedProcCalls: 0,
			expectedWaitTime:  sqsMaxWaitTime,
		},
		{
			name: "Should keep SQS wait time to default",
			fields: fields{
				config: sqsConf,
				sqs: &sqsMock{
					returnMssgs: 1,
				},
				processor: &mockProcessor{},
				logger:    logger,
			},
			expectedErr:       false,
			expectedRecCalls:  1,
			expectedDelCalls:  1,
			expectedProcCalls: 1,
			expectedWaitTime:  defSQSWaitTime,
		},
		{
			name: "Should process N messages",
			fields: fields{
				config: sqsConf,
				sqs: &sqsMock{
					returnMssgs: 5,
				},
				processor: &mockProcessor{},
				logger:    logger,
			},
			expectedErr:       false,
			expectedRecCalls:  1,
			expectedDelCalls:  5,
			expectedProcCalls: 5,
			expectedWaitTime:  defSQSWaitTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consumer := &SQSConsumer{
				config:    tt.fields.config,
				sqs:       tt.fields.sqs,
				processor: tt.fields.processor,
				logger:    tt.fields.logger,
			}

			if err := consumer.readAndProcess(context.Background()); err != nil && !tt.expectedErr {
				t.Fatalf("Expected no error, but got: %v", err.Error())
			}

			if tt.fields.sqs.receiveCalls != tt.expectedRecCalls {
				t.Fatalf("Receive messages calls do not match, expected %d but got %d", tt.expectedRecCalls, tt.fields.sqs.receiveCalls)
			}
			if tt.fields.sqs.deleteCalls != tt.expectedDelCalls {
				t.Fatalf("Delete message calls do not match, expected %d but got %d", tt.expectedDelCalls, tt.fields.sqs.deleteCalls)
			}
			if tt.fields.processor.processCalls != tt.expectedProcCalls {
				t.Fatalf("Process message calls do not match, expected %d but got %d", tt.expectedProcCalls, tt.fields.processor.processCalls)
			}
			if consumer.sqsWaitTime != tt.expectedWaitTime {
				t.Fatalf("Expected SQS wait time to be %d but got %d", tt.expectedWaitTime, consumer.sqsWaitTime)
			}
		})
	}

}
