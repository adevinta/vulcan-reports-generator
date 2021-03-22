/*
Copyright 2021 Adevinta
*/

package upload

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	log "github.com/sirupsen/logrus"

	"github.com/adevinta/vulcan-reports-generator/pkg/model"
)

var (
	testDir  = fmt.Sprintf("/tmp/test-%s", time.Now().String())
	testFile = fmt.Sprintf("%s/testFile", testDir)

	errMock = errors.New("ErrMock")
)

type mockUploadFunc func(*s3manager.UploadInput, ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)

type mockS3Manager struct {
	s3manageriface.UploaderAPI
	mockFunc mockUploadFunc
}

func (m *mockS3Manager) Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	return m.mockFunc(input, options...)
}

// setUp creates a tmp test dir
// and file.
func setUp(t *testing.T) error {
	t.Helper()

	err := os.Mkdir(testDir, os.ModePerm)
	if err != nil {
		t.Fatalf("Error creating test dir: %v", err)
	}

	_, err = os.Create(testFile)
	if err != nil {
		t.Fatalf("Error creating test file: %v", err)
	}

	return nil
}

// tearDown removes tmp test dir.
func tearDown() error {
	return os.RemoveAll(testDir)
}

func TestUpload(t *testing.T) {
	setUp(t)
	defer tearDown()

	testCases := []struct {
		name        string
		input       []model.FileInfo
		mockFunc    mockUploadFunc
		expectedErr error
	}{
		{
			name: "Happy path",
			input: []model.FileInfo{
				model.FileInfo{
					FilePath: testFile,
				},
			},
			mockFunc: func(*s3manager.UploadInput, ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
				// All good.
				return nil, nil
			},
		},
		{
			name: "Should return err opening file",
			input: []model.FileInfo{
				model.FileInfo{
					FilePath: "inexistentTestingFile",
				},
			},
			expectedErr: ErrOpeningFile,
		},
		{
			name: "Should return uploading error",
			input: []model.FileInfo{
				model.FileInfo{
					FilePath: testFile,
				},
			},
			mockFunc: func(*s3manager.UploadInput, ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
				// Return mock error.
				return nil, errMock
			},
			expectedErr: errMock,
		},
	}

	log := log.New()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			uploader := NewS3Uploader(&mockS3Manager{
				mockFunc: tc.mockFunc,
			}, log)

			err := uploader.Upload(tc.input)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("Expected err: %v\nBut got: %v", tc.expectedErr, err)
			}
		})
	}
}
