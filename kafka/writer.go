/**
* Author: JeffreyBool
* Date: 2021/7/12
* Time: 16:41
* Software: GoLand
 */

package kafka

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	tracex "github.com/LabKiko/kiko-gokit/trace"
	"github.com/Shopify/sarama"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Writer interface {
	SendMessage(ctx context.Context, message *ProducerMessage) error
	Errors() <-chan *ProducerError
	Messages() <-chan *ProducerMessage
	Close() (err error)
}

type writer struct {
	opts WriterOpts

	// Atomic flag indicating whether the writer has been closed.
	closed   uint32
	close    chan bool
	errors   chan *ProducerError
	messages chan *ProducerMessage

	// Manages the current batch being aggregated on the writer.
	mutex         sync.Mutex
	syncProducer  sarama.SyncProducer
	asyncProducer sarama.AsyncProducer
}

// NewWriter 初始化
func NewWriter(brokers []string, opts ...WriterOpt) (Writer, error) {
	var (
		err error
	)

	options := newWriterOptions(brokers, opts...)
	config := sarama.NewConfig()
	config.ClientID = options.ServiceName
	config.Net.KeepAlive = 60 * time.Second
	config.Producer.RequiredAcks = sarama.RequiredAcks(options.RequiredAck)
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.MaxMessageBytes = int(sarama.MaxRequestSize - 1)
	config.Version = sarama.V3_0_0_0
	w := &writer{
		opts:     options,
		close:    make(chan bool),
		errors:   make(chan *ProducerError),
		messages: make(chan *ProducerMessage),
	}

	// 异步配置
	if options.Async {
		config.Producer.Retry.Max = options.MaxAttempts
		w.asyncProducer, err = sarama.NewAsyncProducer(options.Brokers, config)
		w.asyncProducer = otelsarama.WrapAsyncProducer(config, w.asyncProducer)
	} else {
		config.Producer.Timeout = 5 * time.Second
		if v := options.ReadTimeout; v > 0 {
			config.Producer.Timeout = time.Duration(v) * time.Second
		}
		w.syncProducer, err = sarama.NewSyncProducer(options.Brokers, config)
		w.syncProducer = otelsarama.WrapSyncProducer(config, w.syncProducer)
	}
	if err != nil {
		return nil, err
	}
	if options.Async {
		go w.eventNotification()
	}

	return w, nil
}

func (w *writer) eventNotification() {
	for {
		select {
		case <-w.close:
			goto close

		case err, ok := <-w.asyncProducer.Errors():
			if !ok {
				goto close
			}

			body, _ := err.Msg.Value.Encode()
			headers := make(map[string]string)
			for _, v := range err.Msg.Headers {
				headers[strings.ToLower(string(v.Key))] = string(v.Value)
			}

			w.opts.Logger.WithFields(map[string]interface{}{
				"topic":     err.Msg.Topic,
				"offset":    err.Msg.Offset,
				"partition": err.Msg.Partition,
				"body":      string(body),
				"header":    headers,
			}).Errorf("producerError: %v", err.Err)

		case msg, ok := <-w.asyncProducer.Successes():
			if !ok {
				goto close
			}

			body, _ := msg.Value.Encode()
			headers := make(map[string]string)
			for _, v := range msg.Headers {
				headers[strings.ToLower(string(v.Key))] = string(v.Value)
			}

			w.opts.Logger.WithFields(map[string]interface{}{
				"topic":     msg.Topic,
				"offset":    msg.Offset,
				"partition": msg.Partition,
				"body":      string(body),
				"header":    headers,
			}).Debug("send msg success")
		}
	}

close:
}

func (w *writer) SendMessage(ctx context.Context, message *ProducerMessage) (err error) {
	var (
		span trace.Span
	)

	if w.isClosed() {
		return io.ErrClosedPipe
	}

	tr := tracex.NewTracer(trace.SpanKindProducer)
	ctx, span = tr.Start(ctx, fmt.Sprintf("KF Producer %s", message.Topic))
	defer span.End()

	msg := ToProducerMessage(message)

	tr.Inject(ctx, otelsarama.NewProducerMessageCarrier(msg))

	if w.opts.Async {
		err = w.asyncSend(ctx, msg)
	} else {
		_, _, err = w.syncSend(ctx, msg)
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return err
}

func (w *writer) syncSend(crx context.Context, msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	if w.syncProducer == nil {
		return 0, 0, errors.New("kafka sync producer client not init")
	}
	return w.syncProducer.SendMessage(msg)

}

func (w *writer) asyncSend(ctx context.Context, msg *sarama.ProducerMessage) (err error) {
	if w.asyncProducer == nil {
		return errors.New("kafka async producer client not init")
	}

	w.asyncProducer.Input() <- msg
	return nil
}

func (w *writer) markClosed() {
	atomic.StoreUint32(&w.closed, 1)
}

func (w *writer) isClosed() bool {
	return atomic.LoadUint32(&w.closed) != 0
}

// Errors returns a read channel of errors that occur during offset management, if
// enabled. By default, errors are logged and not returned over this channel. If
// you want to implement any custom error handling, set your config's
// Consumer.Return.Errors setting to true, and read from this channel.
func (w *writer) Errors() <-chan *ProducerError { return w.errors }

func (w *writer) Messages() <-chan *ProducerMessage { return w.messages }

func (w *writer) Close() (err error) {
	if w.isClosed() {
		return
	}

	close(w.close)
	close(w.errors)
	close(w.messages)
	if w.opts.Async {
		err = w.asyncProducer.Close()
	} else {
		err = w.syncProducer.Close()
	}
	if err != nil {
		return err
	}

	w.markClosed()

	w.opts.Logger.Infof("closed success")
	return nil
}
