/*
Copyright 2021 Adevinta
*/

package upload

import "github.com/adevinta/vulcan-reports-generator/pkg/model"

// Uploader specifies the interface
// for a report's FileInfo uploader.
type Uploader interface {
	Upload(files []model.FileInfo) error
}
