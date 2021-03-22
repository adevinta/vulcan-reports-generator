/*
Copyright 2021 Adevinta
*/

package notify

import (
	"errors"

	"github.com/adevinta/vulcan-reports-generator/pkg/model"
)

var (
	// ErrInvalidConfig indicates that supplied configuration is not valid.
	ErrInvalidConfig = errors.New("Invalid configuration")
	// ErrUnsupportedFmt indicates that the specified format is not supported.
	ErrUnsupportedFmt = errors.New("Unsupported format")
)

// Notifier defines the
// interface for a notifier.
type Notifier interface {
	Notify(subject, mssg string, fmt model.NotifFmt, recipients []string) error
}
