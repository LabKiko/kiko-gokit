/**
* Author: JeffreyBool
* Date: 2021/7/12
* Time: 16:52
* Software: GoLand
 */

package kafka

import (
	"strings"

	"github.com/LabKiko/kiko-gokit/logger"
)

const (
	consumerEvent              = "consumer"
	producerEvent              = "producer"
	defaultProducerServiceName = "go-kafka-producer"
	defaultConsumerServiceName = "go-kafka-consumer"
)

type StartOffset int

const (
	// OffsetNewest stands for the log head offset, i.e. the offset that will be
	// assigned to the next message that will be produced to the partition. You
	// can send this to a client's GetOffset method to get this offset, or when
	// calling ConsumePartition to start consuming new messages.
	OffsetNewest StartOffset = -1
	// OffsetOldest stands for the oldest offset available on the broker for a
	// partition. You can send this to a client's GetOffset method to get this
	// offset, or when calling ConsumePartition to start consuming from the
	// oldest offset that is still available on the broker.
	OffsetOldest StartOffset = -2
)

type ReaderOpts struct {
	ServiceName string
	// The list of broker addresses used to connect to the kafka cluster.
	Brokers []string

	// The topic to read messages from.
	Topic string

	// GroupID holds the optional consumer group id.  If GroupID is specified, then
	// Partition should NOT be specified e.g. 0
	GroupID string

	// StartOffset determines from whence the consumer group should begin
	// consuming when it finds a partition without a committed offset.  If
	// non-zero, it must be set to one of FirstOffset or LastOffset.
	//
	// Default: FirstOffset
	//
	// Only used when Group is set
	StartOffset StartOffset

	CommitInterval int

	Logger logger.Logger
}

func newReaderOptions(brokers []string, topic, group string, opts ...ReaderOpt) *ReaderOpts {
	opt := &ReaderOpts{
		ServiceName: defaultConsumerServiceName,
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     group,
		StartOffset: OffsetOldest,
	}

	for _, o := range opts {
		o(opt)
	}

	if opt.Logger == nil {
		opt.Logger = logger.New(
			logger.WithNamespace("kafka"),
			logger.WithConsole(true),
			logger.WithDisableDisk(true),
			logger.WithFields(map[string]interface{}{
				"app_id": opt.ServiceName,
				"event":  consumerEvent,
				"broker": strings.Join(brokers, ","),
				"topic":  topic,
				"group":  group,
			}),
		)
	}

	InitLogger(opt.Logger)

	return opt
}

type ReaderOpt func(o *ReaderOpts)

func ReaderServiceName(serviceName string) ReaderOpt {
	return func(o *ReaderOpts) {
		o.ServiceName = serviceName
	}
}

func ReaderStartOffset(offset StartOffset) ReaderOpt {
	return func(o *ReaderOpts) {
		o.StartOffset = offset
	}
}

func ReaderCommitInterval(commitInterval int) ReaderOpt {
	return func(o *ReaderOpts) {
		o.CommitInterval = commitInterval
	}
}

func ReaderLogger(logger logger.Logger) ReaderOpt {
	return func(o *ReaderOpts) {
		o.Logger = logger
	}
}

type RequiredAck int16

const (
	// NoResponse doesn't send any response, the TCP ACK is all you get.
	NoResponse RequiredAck = 0
	// WaitForLocal waits for only the local commit to succeed before responding.
	WaitForLocal RequiredAck = 1
	// WaitForAll waits for all in-sync replicas to commit before responding.
	// The minimum number of in-sync replicas is configured on the broker via
	// the `min.insync.replicas` configuration key.
	WaitForAll RequiredAck = -1
)

type WriterOpts struct {
	ServiceName string
	Brokers     []string
	// Limit on how many attempts will be made to deliver a message.
	//
	// The default is to try at most 10 times.
	MaxAttempts int

	// Number of acknowledges from partition replicas required before receiving
	// a response to a produce request. The default is -1, which means to wait for
	// all replicas, and a value above 0 is required to indicate how many replicas
	// should acknowledge a message to be considered successful.
	//
	// This version of kafka-go (v0.3) does not support 0 required acks, due to
	// some internal complexity implementing this with the Kafka protocol. If you
	// need that functionality specifically, you'll need to upgrade to v0.4.
	RequiredAck RequiredAck

	ReadTimeout int

	Async bool

	Logger logger.Logger
}

type WriterOpt func(o *WriterOpts)

func newWriterOptions(brokers []string, opts ...WriterOpt) WriterOpts {
	opt := WriterOpts{
		ServiceName: defaultProducerServiceName,
		Brokers:     brokers,
		RequiredAck: WaitForAll,
		MaxAttempts: 3,
	}
	for _, o := range opts {
		o(&opt)
	}
	if opt.Logger == nil {
		opt.Logger = logger.New(
			logger.WithNamespace("kafka"),
			logger.WithConsole(true),
			logger.WithDisableDisk(true),
			logger.WithFields(map[string]interface{}{
				"app_id": opt.ServiceName,
				"event":  producerEvent,
				"broker": strings.Join(brokers, ","),
			}),
		)
	}

	InitLogger(opt.Logger)

	return opt
}

func WriterServiceName(serviceName string) WriterOpt {
	return func(o *WriterOpts) {
		o.ServiceName = serviceName
	}
}

func WriterMaxAttempts(num int) WriterOpt {
	return func(o *WriterOpts) {
		o.MaxAttempts = num
	}
}

func WriterRequiredAck(requiredAck RequiredAck) WriterOpt {
	return func(o *WriterOpts) {
		o.RequiredAck = requiredAck
	}
}

func WriterReadTimeout(readTimeout int) WriterOpt {
	return func(o *WriterOpts) {
		o.ReadTimeout = readTimeout
	}
}

func WriterAsync(async bool) WriterOpt {
	return func(o *WriterOpts) {
		o.Async = async
	}
}

func WriterLogger(logger logger.Logger) WriterOpt {
	return func(o *WriterOpts) {
		o.Logger = logger
	}
}
