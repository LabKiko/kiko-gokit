package agent

import (
	"time"

	"github.com/LabKiko/kiko-gokit/nightingale-agent/types"
)

type Input interface {
	// Init initializes the input.
	Init() error
	// Gather gathers the data.
	Gather() ([]*types.Sample, error)
	// GetInterval returns the interval of the input.
	GetInterval() time.Duration
	// Prefix returns the prefix of the input.
	Prefix() string
	// String returns the name of the input.
	String() string
	// Close closes the input.
	Close() error
}
