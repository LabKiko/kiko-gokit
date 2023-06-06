package kafka

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/LabKiko/kiko-gokit/logger"
	tracex "github.com/LabKiko/kiko-gokit/trace"
	"github.com/Shopify/sarama"
	"go.opentelemetry.io/contrib/instrumentation/github.com/Shopify/sarama/otelsarama"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

type Reader interface {
	FetchMessage(ctx context.Context, handler Handler) error
	CommitMessage(ctx context.Context, session sarama.ConsumerGroupSession, message *ConsumerMessage) error
	Close() error
}

type reader struct {
	opts *ReaderOpts

	// mutable fields of the reader (synchronized on the mutex)
	mutex    sync.Mutex
	ctx      context.Context
	Cancel   context.CancelFunc
	close    chan bool
	closed   int32
	consumer sarama.ConsumerGroup

	handler Handler
}

type Handler func(ctx context.Context, session sarama.ConsumerGroupSession, message *ConsumerMessage) error

func NewReader(brokers []string, topic, group string, opts ...ReaderOpt) (Reader, error) {
	reader := &reader{
		close: make(chan bool),
	}

	reader.opts = newReaderOptions(brokers, topic, group, opts...)
	config := sarama.NewConfig()
	config.Version = sarama.V3_0_0_0
	config.ClientID = reader.opts.ServiceName
	config.Consumer.Offsets.Initial = int64(reader.opts.StartOffset)
	config.Consumer.Return.Errors = true
	if reader.opts.CommitInterval > 0 { // 设置自动提交offset间隔,默认10s
		config.Consumer.Offsets.AutoCommit.Enable = true
		config.Consumer.Offsets.AutoCommit.Interval = time.Duration(reader.opts.CommitInterval) * time.Second
	} else {
		config.Consumer.Offsets.AutoCommit.Enable = true
		config.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second
	}
	config.Producer.MaxMessageBytes = int(sarama.MaxRequestSize - 1) // 1M
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	client, err := sarama.NewConsumerGroup(reader.opts.Brokers, group, config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	reader.ctx = ctx
	reader.Cancel = cancel
	reader.consumer = client

	go func() {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, []string{reader.opts.Topic}, otelsarama.WrapConsumerGroupHandler(reader)); err != nil {
				reader.opts.Logger.Errorf("consume error: %v", err)
			}

			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
		}
	}()

	go reader.eventNotification()

	return reader, nil
}

func (r *reader) eventNotification() {
	for {
		select {
		case <-r.close:
			goto close

		case err, ok := <-r.consumer.Errors():
			if !ok {
				goto close
			}

			r.opts.Logger.Errorf("consumer error: %v", err)
		}
	}

close:
}

func (r *reader) FetchMessage(ctx context.Context, handler Handler) error {
	if r.isClosed() {
		return io.EOF
	}

	r.handler = handler
	return nil
}

func (r *reader) CommitMessage(ctx context.Context, session sarama.ConsumerGroupSession, message *ConsumerMessage) error {
	session.MarkMessage(encodedConsumerMessage(message), "")
	return nil
}

func (r *reader) markClosed() {
	atomic.StoreInt32(&r.closed, 1)
}

func (r *reader) isClosed() bool {
	return atomic.LoadInt32(&r.closed) != 0
}

func (r *reader) Close() error {
	if r.isClosed() {
		return nil
	}

	r.Cancel()
	close(r.close)
	r.markClosed()
	err := r.consumer.Close()
	if err != nil {
		return err
	}

	r.opts.Logger.Debugf("closed success")
	return nil
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (r *reader) Setup(session sarama.ConsumerGroupSession) error {
	r.opts.Logger.Debugf("Consume Setup GenerationID: %d, MemberID: %s", session.GenerationID(), session.MemberID())
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (r *reader) Cleanup(session sarama.ConsumerGroupSession) error {
	r.opts.Logger.Debugf("Consume Cleanup GenerationID: %d, MemberID: %s", session.GenerationID(), session.MemberID())
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (r *reader) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for {
		select {
		case <-r.close:
			return nil
		case <-session.Context().Done():
			return nil
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			if r.handler == nil {
				r.opts.Logger.Error("ignore, unregistered handler")
				continue
			}

			err := r.Handler(message, session, claim)
			if err != nil {
				logger.Error(err)
			}
		}
	}
}

// Handler handler message
func (r *reader) Handler(msg *sarama.ConsumerMessage, session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	var (
		span trace.Span
	)

	tr := tracex.NewTracer(trace.SpanKindConsumer)

	// Extract tracing info from message
	ctx := tr.Extract(context.Background(), otelsarama.NewConsumerMessageCarrier(msg))
	bags := baggage.FromContext(ctx)
	spanCtx := trace.SpanContextFromContext(ctx)
	ctx = baggage.ContextWithBaggage(ctx, bags)

	ctx, span = tr.Start(trace.ContextWithRemoteSpanContext(ctx, spanCtx), fmt.Sprintf("Kafka Consumer %s", msg.Topic), trace.WithAttributes(
		semconv.MessagingOperationProcess,
	))
	defer span.End()

	err := r.handler(ctx, session, decodeConsumerMessage(msg))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}
