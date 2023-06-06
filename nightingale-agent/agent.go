package agent

import (
	"context"
	"os"

	"github.com/LabKiko/kiko-gokit/logger"
)

var hostname = "unknown"

func init() {
	hostname, _ = os.Hostname()
}

type Agent interface {
	// Start starts the agent
	Start() error
	// Flush flushes the metrics
	Flush() error
	// Stop stops the agent
	Stop() error
}

type defaultAgent struct {
	opt      Options
	ctx      context.Context
	cancelFn func()
	readers  map[string]Reader
}

// New creates a new Agent
func New(endpoint string, opts ...Option) (Agent, error) {
	writer, err := NewRemoteWriter(endpoint)
	if err != nil {
		return nil, err
	}

	_op := Options{}
	for _, opt := range opts {
		opt(&_op)
	}
	op := Options{
		Output: NewOutput(writer, WithDebug(_op.Debug)),
	}
	for _, opt := range opts {
		opt(&op)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &defaultAgent{
		opt:      op,
		ctx:      ctx,
		cancelFn: cancel,
		readers:  make(map[string]Reader),
	}, nil
}

// Start starts the agent
func (agent *defaultAgent) Start() error {
	for _, input := range agent.opt.Inputs {
		if err := input.Init(); err != nil {
			logger.Errorf("failed to initialize input: %s, err: %s", input.String(), err)
			continue
		}

		reader := NewReader(agent.opt.buildReader()...)
		if err := reader.Init(input, agent.opt.Output); err != nil {
			logger.Errorf("failed to initialize reader: %s, err: %s", input.String(), err)
			continue
		}

		if err := reader.Start(); err != nil {
			logger.Errorf("failed to start reader: %s, err: %s", input.String(), err)
			continue
		}

		agent.readers[input.String()] = reader
		logger.Infof("reader %s started", input.String())
	}

	logger.Infof("metrics agent started")
	return nil
}

// Flush flushes the metrics
func (agent *defaultAgent) Flush() error {
	for _, reader := range agent.readers {
		if err := reader.Flush(); err != nil {
			logger.Errorf("failed to flush reader %s, err: %s", reader.String(), err)
		}
	}

	return nil
}

// Stop stops the agent
func (agent *defaultAgent) Stop() error {
	if agent.cancelFn != nil {
		agent.cancelFn()
	}

	for _, reader := range agent.readers {
		if err := reader.Stop(); err != nil {
			logger.Errorf("failed to stop reader %s, err: %s", reader.String(), err)
		}
	}

	logger.Infof("metrics agent stopped")
	return nil
}
