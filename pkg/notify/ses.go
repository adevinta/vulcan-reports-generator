/*
Copyright 2021 Adevinta
*/

package notify

import (
	"net/mail"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"

	"github.com/adevinta/vulcan-reports-generator/pkg/model"
)

const (
	utf8 = "UTF-8"
)

type sesNotifier struct {
	cfg    SESConfig
	sesSvc sesiface.SESAPI
}

type SESConfig struct {
	Region string
	From   string
	CC     []string
}

func NewSESNotifier(cfg SESConfig, sesSvc sesiface.SESAPI) (*sesNotifier, error) {
	if !isValidConfig(cfg) {
		return nil, ErrInvalidConfig
	}

	return &sesNotifier{
		cfg:    cfg,
		sesSvc: sesSvc,
	}, nil
}

func (n *sesNotifier) Notify(subject, mssg string, fmt model.NotifFmt, recipients []string) error {
	input, err := n.buildInput(subject, mssg, fmt, recipients)
	if err != nil {
		return err
	}

	_, err = n.sesSvc.SendEmail(input)
	if err != nil {
		return err
	}

	return nil
}

func (n *sesNotifier) buildInput(subject, mssg string, fmt model.NotifFmt, recipients []string) (*ses.SendEmailInput, error) {
	content := &ses.Content{
		Charset: aws.String(utf8),
		Data:    aws.String(mssg),
	}

	var body *ses.Body
	switch fmt {
	case model.NotifFmtHTML:
		body = &ses.Body{Html: content}
	case model.NotifFmtText:
		body = &ses.Body{Text: content}
	default:
		return nil, ErrUnsupportedFmt
	}

	return &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: stringSliceToAWSString(n.cfg.CC),
			ToAddresses: stringSliceToAWSString(recipients),
		},
		Message: &ses.Message{
			Body: body,
			Subject: &ses.Content{
				Charset: aws.String(utf8),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(n.cfg.From),
	}, nil
}

func isValidConfig(cfg SESConfig) bool {
	if _, err := mail.ParseAddress(cfg.From); err != nil ||
		cfg.From == "" || cfg.Region == "" {
		return false
	}

	// Validate CC.
	for _, cc := range cfg.CC {
		if _, err := mail.ParseAddress(cc); err != nil {
			return false
		}
	}

	return true
}

func stringSliceToAWSString(strs []string) (awsStr []*string) {
	for _, str := range strs {
		awsStr = append(awsStr, aws.String(str))
	}
	return
}
