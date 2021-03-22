/*
Copyright 2021 Adevinta
*/

package notify

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"

	"github.com/adevinta/vulcan-reports-generator/pkg/model"
)

var (
	errMock = errors.New("ErrMock")

	mockCfg = SESConfig{
		Region: "eu-west-1",
		From:   "me@me.com",
	}
)

type mockSendEmailFunc func(*ses.SendEmailInput) (*ses.SendEmailOutput, error)

type mockSESAPI struct {
	sesiface.SESAPI
	mockFunc mockSendEmailFunc
}

func (m *mockSESAPI) SendEmail(input *ses.SendEmailInput) (*ses.SendEmailOutput, error) {
	return m.mockFunc(input)
}

func TestNotify(t *testing.T) {
	type input struct {
		subject    string
		mssg       string
		fmt        model.NotifFmt
		recipients []string
	}

	testCases := []struct {
		name        string
		input       input
		mockFunc    mockSendEmailFunc
		expectedErr error
	}{
		{
			name: "Happy path",
			input: input{
				subject:    "An important subject",
				mssg:       "An important mssg",
				fmt:        model.NotifFmtText,
				recipients: []string{"tom@somewhere.com"},
			},
			mockFunc: func(*ses.SendEmailInput) (*ses.SendEmailOutput, error) {
				// All good.
				return nil, nil
			},
		},
		{
			name: "Should return err due to SendEmail err",
			input: input{
				subject:    "An important subject",
				mssg:       "An important mssg",
				fmt:        model.NotifFmtHTML,
				recipients: []string{"tom@somewhere.com"},
			},
			mockFunc: func(*ses.SendEmailInput) (*ses.SendEmailOutput, error) {
				// Return mock err.
				return nil, errMock
			},
			expectedErr: errMock,
		},
		{
			name: "Should return err due to invalid fmt",
			input: input{
				subject:    "An important subject",
				mssg:       "An important mssg",
				fmt:        100, // Invalid fmt.
				recipients: []string{"tom@somewhere.com"},
			},
			expectedErr: ErrUnsupportedFmt,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			notifier := &sesNotifier{
				cfg: mockCfg,
				sesSvc: &mockSESAPI{
					mockFunc: tc.mockFunc,
				},
			}

			err := notifier.Notify(tc.input.mssg, tc.input.subject, tc.input.fmt, tc.input.recipients)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("Expected err: %v\nBut got: %v", tc.expectedErr, err)
			}
		})
	}
}

func TestParseConfig(t *testing.T) {
	testCases := []struct {
		name     string
		input    SESConfig
		expected bool
	}{
		{
			name: "Happy path",
			input: SESConfig{
				Region: "eu",
				From:   "from@awe.com",
				CC:     []string{"cc@awe.com"},
			},
			expected: true,
		},
		{
			name: "Should return true, cfg with no CC",
			input: SESConfig{
				Region: "eu",
				From:   "from@awe.com",
			},
			expected: true,
		},
		{
			name: "Should return false due to no From",
			input: SESConfig{
				Region: "eu",
			},
			expected: false,
		},
		{
			name: "Should return false due to invalid From",
			input: SESConfig{
				Region: "eu",
				From:   "invalidEmail.com",
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := isValidConfig(tc.input)
			if isValid != tc.expected {
				t.Fatalf("Expected %v, but got %v", tc.expected, isValid)
			}
		})
	}
}
