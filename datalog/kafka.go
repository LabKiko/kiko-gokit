package datalog

import (
	"context"
	"encoding/json"

	"github.com/LabKiko/kiko-gokit/kafka"
)

type Kafka struct {
	opts   *config
	writer kafka.Writer
}

func NewKafka(opts *config) (*Kafka, error) {
	cli := &Kafka{
		opts: opts,
	}
	writer, err := kafka.NewWriter(opts.kafkaOpts.brokers)
	if err != nil {
		return nil, err
	}

	cli.writer = writer

	return cli, nil
}

func (w *Kafka) Write(ctx context.Context, event *Event, metadata Metadata) error {
	bytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	return w.writer.SendMessage(ctx, &kafka.ProducerMessage{
		Topic: "backenddot." + w.opts.appId + "-" + event.Name,
		Value: bytes,
	})
}

func (w *Kafka) Flush() error {
	return nil
}

func (w *Kafka) Close() error {
	return w.writer.Close()
}
