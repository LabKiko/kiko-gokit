package logger

import (
	"errors"
)

var (
	// ErrLogPathNotSet is an error that indicates the log path is not set.
	ErrLogPathNotSet = errors.New("log path must be set")
)
