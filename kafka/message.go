/**
* Author: JeffreyBool
* Date: 2021/7/13
* Time: 01:11
* Software: GoLand
 */

package kafka

import (
	"time"

	"github.com/Shopify/sarama"
	uuid "github.com/satori/go.uuid"
)

// RecordHeader stores key and value for a record header
type RecordHeader struct {
	Key   []byte
	Value []byte
}

// ConsumerMessage encapsulates a Kafka message returned by the consumer.
type ConsumerMessage struct {
	Key, Value []byte
	Topic      string
	Partition  int32
	Offset     int64

	Headers        []*RecordHeader // only set if kafka is version 0.11+
	Timestamp      time.Time       // only set if kafka is version 0.10+, inner message timestamp
	BlockTimestamp time.Time       // only set if kafka is version 0.10+, outer (compressed) block timestamp
}

func decodeConsumerMessage(message *sarama.ConsumerMessage) *ConsumerMessage {
	headers := make([]*RecordHeader, len(message.Headers))
	for i, h := range message.Headers {
		headers[i] = &RecordHeader{Key: h.Key, Value: h.Value}
	}
	m := &ConsumerMessage{
		Key:            message.Key,
		Value:          message.Value,
		Topic:          message.Topic,
		Partition:      message.Partition,
		Offset:         message.Offset,
		Timestamp:      message.Timestamp,
		BlockTimestamp: message.BlockTimestamp,
		Headers:        headers,
	}
	return m
}

func encodedConsumerMessage(message *ConsumerMessage) *sarama.ConsumerMessage {
	var headers = make([]*sarama.RecordHeader, 0, len(message.Headers))
	for _, header := range message.Headers {
		headers = append(headers, &sarama.RecordHeader{
			Key:   header.Key,
			Value: header.Value,
		})
	}

	return &sarama.ConsumerMessage{
		Headers:        headers,
		Timestamp:      message.Timestamp,
		BlockTimestamp: message.BlockTimestamp,
		Key:            message.Key,
		Value:          message.Value,
		Topic:          message.Topic,
		Partition:      message.Partition,
		Offset:         message.Offset,
	}
}

type ProducerMessage struct {
	Topic string // The Kafka topic for this message.
	// The partitioning key for this message. Pre-existing Encoders include
	// StringEncoder and ByteEncoder.
	Key string
	// The actual message to store in Kafka. Pre-existing Encoders include
	// StringEncoder and ByteEncoder.
	Value []byte

	// This field is used to hold arbitrary data you wish to include so it
	// will be available when receiving on the Successes and Errors channels.
	// Sarama completely ignores this field and is only to be used for
	// pass-through data.
	Metadata interface{}

	// Below this point are filled in by the producer as the message is processed

	// Offset is the offset of the message stored on the broker. This is only
	// guaranteed to be defined if the message was successfully delivered and
	// RequiredAcks is not NoResponse.
	Offset int64
	// Partition is the partition that the message was sent to. This is only
	// guaranteed to be defined if the message was successfully delivered.
	Partition int32
	// Timestamp is the timestamp assigned to the message by the broker. This
	// is only guaranteed to be defined if the message was successfully
	// delivered, RequiredAcks is not NoResponse, and the Kafka broker is at
	// least version 0.10.0.
	Timestamp time.Time

	// MessageID
	MessageID string
}

// ProducerError is the type of error generated when the producer fails to deliver a message.
// It contains the original ProducerMessage as well as the actual error value.
type ProducerError struct {
	Msg *ProducerMessage
	Err error
}

func decodeProducerMessage(message *sarama.ProducerMessage) *ProducerMessage {
	key, _ := message.Key.Encode()
	value, _ := message.Value.Encode()
	return &ProducerMessage{
		Topic:     message.Topic,
		Key:       string(key),
		Value:     value,
		Metadata:  message.Metadata,
		Offset:    message.Offset,
		Partition: message.Partition,
		Timestamp: message.Timestamp,
	}
}

func encodedProducerMessage(message *ProducerMessage) *sarama.ProducerMessage {
	msg := &sarama.ProducerMessage{}
	now := time.Now()
	msg.Topic = message.Topic
	if message.Partition <= 0 {
		msg.Partition = int32(-1)
	}
	if message.Key == "" {
		u1 := uuid.NewV4()
		msg.Key = sarama.ByteEncoder(u1.Bytes())
	} else {
		msg.Key = sarama.StringEncoder(message.Key)
	}
	msg.Value = sarama.ByteEncoder(message.Value)
	if message.Timestamp.IsZero() {
		msg.Timestamp = now
	} else {
		msg.Timestamp = message.Timestamp
	}
	return msg
}
