/**
* Author: JeffreyBool
* Date: 2021/11/10
* Time: 22:06
* Software: GoLand
 */

package kafka

import (
	"context"
	"fmt"
	"testing"
	"time"

	tracex "github.com/LabKiko/kiko-gokit/trace"
	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

const (
	addr  = "10.130.12.10:9092"
	topic = "infra.datalog.biz-test.data"
	group = "datalog.biz"
)

func Test_reader_FetchMessage(t *testing.T) {
	_, err := tracex.New(tracex.WithName("kafka-reader"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = InitReader([]string{addr}, topic, group)
	if err != nil {
		t.Fatal(err)
	}
	defer _reader.Close()

	err = _reader.FetchMessage(context.Background(), func(ctx context.Context, session sarama.ConsumerGroupSession, message *ConsumerMessage) error {
		HandlerMsgBiz(ctx, message)
		return _reader.CommitMessage(ctx, session, message)
	})
	if err != nil {
		assert.Error(t, err)
	}

	select {}
}

func HandlerMsgBiz(ctx context.Context, message *ConsumerMessage) {
	tr := tracex.NewTracer(trace.SpanKindInternal)
	var span trace.Span
	ctx, span = tr.Start(ctx, "HandlerMsgBiz")
	defer span.End()

	time.Sleep(time.Second * 3)
	fmt.Println("===========================>", string(message.Value))
	fmt.Println("trace:", tracex.ExtractTraceId(ctx))
}
