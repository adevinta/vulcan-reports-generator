/*
Copyright 2021 Adevinta
*/

package upload

import (
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	log "github.com/sirupsen/logrus"

	"github.com/adevinta/vulcan-reports-generator/pkg/model"
)

var (
	// ErrOpeningFile indicates an error opening a file to upload.
	ErrOpeningFile = errors.New("Error opening file")
)

type s3Uploader struct {
	uploadManager s3manageriface.UploaderAPI
	log           *log.Logger
}

// NewS3Uploader builds and returns a new uploader for AWS S3.
func NewS3Uploader(uploadManager s3manageriface.UploaderAPI, log *log.Logger) Uploader {
	return &s3Uploader{
		uploadManager: uploadManager,
		log:           log,
	}
}

func (u *s3Uploader) Upload(files []model.FileInfo) error {
	for _, fi := range files {
		f, err := os.Open(fi.FilePath)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrOpeningFile, err)
		}

		contentType := mime.TypeByExtension(filepath.Ext(f.Name()))

		u.log.WithFields(log.Fields{
			"bucket": fi.DestBucket,
			"key":    fi.DestPath,
		}).Trace("Uploading file")
		_, err = u.uploadManager.Upload(&s3manager.UploadInput{
			Bucket:      aws.String(fi.DestBucket),
			Key:         aws.String(fi.DestPath),
			Body:        f,
			ContentType: aws.String(contentType),
		})
		if err != nil {
			return err
		}

		f.Close()
	}
	return nil
}
