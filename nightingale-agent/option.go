package agent

import (
	"time"
)

type Option func(o *Options)

type Options struct {
	Debug         bool          // debug mode
	Inputs        []Input       // inputs to use
	Output        Output        // output to use
	ReaderOptions ReaderOptions // reader options
}

func (o *Options) buildReader() []ReaderOption {
	opts := make([]ReaderOption, 0)
	if o.ReaderOptions.Tags != nil {
		opts = append(opts, WithTags(o.ReaderOptions.Tags))
	}
	if o.ReaderOptions.BatchSize != 0 {
		opts = append(opts, WithBatchSize(o.ReaderOptions.BatchSize))
	}
	if o.ReaderOptions.QueueSize != 0 {
		opts = append(opts, WithQueueSize(o.ReaderOptions.QueueSize))
	}
	if o.ReaderOptions.Interval != 0 {
		opts = append(opts, WithInterval(o.ReaderOptions.Interval))
	}
	if o.ReaderOptions.OutputInterval != 0 {
		opts = append(opts, WithOutputInterval(o.ReaderOptions.OutputInterval))
	}
	return opts
}

// WithOutputDebug sets a debug mode for the output.
func WithOutputDebug(v bool) Option {
	return func(o *Options) {
		o.Debug = v
	}
}

// WithInput sets an input to use.
func WithInput(inputs ...Input) Option {
	return func(o *Options) {
		o.Inputs = append(o.Inputs, inputs...)
	}
}

// WithOutput sets an output to use.
func WithOutput(output Output) Option {
	return func(o *Options) {
		o.Output = output
	}
}

// WithReaderTags sets tags to use for the reader.
func WithReaderTags(tags map[string]string) Option {
	return func(o *Options) {
		o.ReaderOptions.Tags = tags
	}
}

// WithReaderBatchSize sets a batch size for the reader.
func WithReaderBatchSize(batchSize int) Option {
	return func(o *Options) {
		o.ReaderOptions.BatchSize = batchSize
	}
}

// WithReaderQueueSize sets a queue size for the reader.
func WithReaderQueueSize(queueSize int) Option {
	return func(o *Options) {
		o.ReaderOptions.QueueSize = queueSize
	}
}

// WithReaderInterval sets an interval for the reader.
func WithReaderInterval(interval time.Duration) Option {
	return func(o *Options) {
		o.ReaderOptions.Interval = interval
	}
}

// WithReaderOutputInterval sets an interval for the output.
func WithReaderOutputInterval(interval time.Duration) Option {
	return func(o *Options) {
		o.ReaderOptions.OutputInterval = interval
	}
}

type RemoteWriterOption func(o *RemoteWriterOptions)

type RemoteWriterOptions struct {
	Endpoint              string        // endpoint of the remote writer
	BasicUsername         string        // basic auth username
	BasicPassword         string        // basic auth password
	DialTimeout           time.Duration // timeout for the dial
	ResponseHeaderTimeout time.Duration // timeout for the response header
	MaxIdleConnsPerHost   int           // max idle conns per host
}

// WithBasicUsername sets a basic auth username to use.
func WithBasicUsername(username string) RemoteWriterOption {
	return func(o *RemoteWriterOptions) {
		o.BasicUsername = username
	}
}

// WithBasicPassword sets a basic auth password to use.
func WithBasicPassword(basicPassword string) RemoteWriterOption {
	return func(o *RemoteWriterOptions) {
		o.BasicPassword = basicPassword
	}
}

// WithResponseHeaderTimeout sets a timeout for the response header.
func WithResponseHeaderTimeout(timeout time.Duration) RemoteWriterOption {
	return func(o *RemoteWriterOptions) {
		o.ResponseHeaderTimeout = timeout
	}
}

// WithDialTimeout sets a timeout for the dial.
func WithDialTimeout(dialTimeout time.Duration) RemoteWriterOption {
	return func(o *RemoteWriterOptions) {
		o.DialTimeout = dialTimeout
	}
}

// WithMaxIdleConnsPerHost sets a max idle conns per host to use.
func WithMaxIdleConnsPerHost(v int) RemoteWriterOption {
	return func(o *RemoteWriterOptions) {
		o.MaxIdleConnsPerHost = v
	}
}

type ReaderOption func(o *ReaderOptions)

type ReaderOptions struct {
	Tags           map[string]string // tags to use
	QueueSize      int               // queue size for the reader
	Interval       time.Duration     // interval for the reader
	BatchSize      int               // batch size for the reader
	OutputInterval time.Duration     // interval for the output
}

// WithTags sets tags to use for the reader.
func WithTags(tags map[string]string) ReaderOption {
	return func(o *ReaderOptions) {
		o.Tags = tags
	}
}

// WithBatchSize sets a batch size for the reader.
func WithBatchSize(batchSize int) ReaderOption {
	return func(o *ReaderOptions) {
		o.BatchSize = batchSize
	}
}

// WithQueueSize sets a queue size for the reader.
func WithQueueSize(queueSize int) ReaderOption {
	return func(o *ReaderOptions) {
		o.QueueSize = queueSize
	}
}

// WithInterval sets an interval for the reader.
func WithInterval(interval time.Duration) ReaderOption {
	return func(o *ReaderOptions) {
		o.Interval = interval
	}
}

// WithOutputInterval sets an interval for the output.
func WithOutputInterval(interval time.Duration) ReaderOption {
	return func(o *ReaderOptions) {
		o.OutputInterval = interval
	}
}

type OutputOption func(o *OutputOptions)

type OutputOptions struct {
	Debug        bool          // debug mode
	Timeout      time.Duration // timeout for the output
	RemoteWriter RemoteWriter  // Prometheus Protocol remote writer to use
}

// WithDebug sets a debug mode for the output.
func WithDebug(v bool) OutputOption {
	return func(o *OutputOptions) {
		o.Debug = v
	}
}

// WithTimeout sets a timeout for the output.
func WithTimeout(timeout time.Duration) OutputOption {
	return func(o *OutputOptions) {
		o.Timeout = timeout
	}
}
