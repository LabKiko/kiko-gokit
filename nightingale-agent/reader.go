package agent

import (
	"context"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"github.com/LabKiko/kiko-gokit/logger"
	"github.com/LabKiko/kiko-gokit/nightingale-agent/types"
	"github.com/prometheus/common/model"
)

const agentHostnameLabelKey model.LabelName = "agent_hostname"

var metricReplacer = strings.NewReplacer("-", "_", ".", "_", " ", "_", "'", "_", "\"", "_")

type Reader interface {
	Init(input Input, output Output) error
	Start() error
	Stop() error
	Flush() error
	String() string
}

type defaultReader struct {
	name   string
	input  Input
	output Output
	flush  chan struct{}
	queue  chan *types.Sample
	opt    *ReaderOptions
	closed uint32
	ctx    context.Context
	cancel context.CancelFunc
}

// NewReader returns a new Reader.
func NewReader(opts ...ReaderOption) Reader {
	op := &ReaderOptions{
		Interval:       time.Second * 30,
		BatchSize:      2000,
		QueueSize:      10000,
		OutputInterval: time.Second * 60,
	}
	for _, opt := range opts {
		opt(op)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &defaultReader{
		opt:    op,
		ctx:    ctx,
		cancel: cancel,
		flush:  make(chan struct{}, 1),
		queue:  make(chan *types.Sample, op.QueueSize),
	}
}

// Init initializes the reader.
func (reader *defaultReader) Init(input Input, output Output) error {
	reader.input = input
	reader.output = output
	reader.name = reader.input.String()
	return nil
}

// Start starts the reader.
func (reader *defaultReader) Start() error {
	go reader.Consumer()
	go reader.Collector()
	return nil
}

// Consumer consumes the samples.
func (reader *defaultReader) Consumer() {
	ticker := time.NewTicker(reader.opt.OutputInterval)
	defer ticker.Stop()

	batch := reader.opt.BatchSize
	series := make([]*types.Sample, 0, batch)

	var count int
	for {
		select {
		case _, ok := <-reader.flush:
			if !ok {
				goto Close
			}

			if len(series) > 0 {
				if err := reader.output.Write(context.Background(), series); err != nil {
					logger.Errorf("%s: failed to write err: %s", reader.name, err)
				}

				logger.Debugf("input: %s, flush write samples len: %d", reader.name, len(series))
				// reset the pool
				count = 0
				series = make([]*types.Sample, 0, batch)
			}
		case <-reader.ctx.Done():
			goto Close
		case sample, ok := <-reader.queue:
			if !ok {
				goto Close
			}
			if sample == nil {
				continue
			}

			series = append(series, sample)
			count++
			if count < batch {
				continue
			}

			if err := reader.output.Write(context.Background(), series); err != nil {
				logger.Errorf("%s: failed to write err: %s", reader.name, err)
			}

			// reset the pool
			count = 0
			series = make([]*types.Sample, 0, batch)
		case <-ticker.C:
			if len(series) > 0 {
				if err := reader.output.Write(context.Background(), series); err != nil {
					logger.Errorf("%s: failed to write err: %s", reader.name, err)
				}

				// reset the pool
				count = 0
				series = make([]*types.Sample, 0, batch)
			}
		}
	}

Close:
	logger.Infof("input: %s, consumer stopped", reader.name)
}

func (reader *defaultReader) Collector() {
	interval := reader.input.GetInterval()
	if interval == 0 {
		interval = reader.opt.Interval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-reader.ctx.Done():
			goto Close
		case <-ticker.C:
			if err := reader.Gather(); err != nil {
				logger.Errorf("%s: failed to gather metrics: %s", reader.name, err)
			}
		}
	}

Close:

	logger.Infof("input: %s collector stopped", reader.name)
}

func (reader *defaultReader) Gather() error {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("%s: gather metrics panic: %s, stack: %s", reader.name, r, string(debug.Stack()))
		}
	}()

	samples, err := reader.input.Gather()
	if err != nil {
		return err
	}

	size := len(samples)
	if size == 0 {
		return nil
	}

	now := time.Now()
	for _, sample := range samples {
		if sample == nil {
			continue
		}

		if sample.Timestamp.IsZero() {
			sample.Timestamp = now
		}

		sample.Metric = metricReplacer.Replace(sample.Metric)
		if len(reader.input.Prefix()) > 0 {
			sample.Metric = reader.input.Prefix() + "_" + metricReplacer.Replace(sample.Metric)
		}

		// add label: agent_name=<name>
		sample.Labels[agentHostnameLabelKey] = model.LabelValue(hostname)

		// add default agent tag
		for key, val := range reader.opt.Tags {
			k := model.LabelName(key)
			v := model.LabelValue(val)
			if !k.IsValid() {
				logger.Errorf("%s: invalid tag key: %s", reader.name, key)
				continue
			}
			if !v.IsValid() {
				logger.Errorf("%s: invalid tag value: %s", reader.name, val)
				continue
			}
			sample.Labels[k] = v
		}

		// write to remote write queue
		if reader.isClosed() {
			return nil
		}

		reader.queue <- sample
	}

	return nil
}

func (reader *defaultReader) markClosed() {
	atomic.StoreUint32(&reader.closed, 1)
}

func (reader *defaultReader) isClosed() bool {
	return atomic.LoadUint32(&reader.closed) != 0
}

func (reader *defaultReader) Flush() error {
	reader.flush <- struct{}{}
	logger.Infof("input: %s, flush", reader.name)
	return nil
}

// Stop stops the reader.
func (reader *defaultReader) Stop() error {
	if reader.isClosed() {
		return nil
	}
	reader.markClosed()

	// sync write the remaining samples
	reader.Flush()
	time.Sleep(time.Second)

	err := reader.input.Close()
	if err != nil {
		logger.Errorf("failed to close input: %s", err)
	}

	if reader.cancel != nil {
		reader.cancel()
	}

	close(reader.queue)
	close(reader.flush)
	logger.Infof("input: %s stopped", reader.name)
	return nil
}

func (reader *defaultReader) String() string {
	return reader.name
}
