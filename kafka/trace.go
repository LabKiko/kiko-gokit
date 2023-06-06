/**
* Author: JeffreyBool
* Date: 2021/7/16
* Time: 01:14
* Software: GoLand
 */

package kafka

import (
	"time"

	"github.com/Shopify/sarama"
	uuid "github.com/satori/go.uuid"
)

func ToProducerMessage(message *ProducerMessage) (msg *sarama.ProducerMessage) {
	var (
		now = time.Now()
		key sarama.Encoder
	)

	if message.MessageID == "" {
		message.MessageID = uuid.NewV4().String()
	}

	if message.Key == "" {
		u1 := uuid.NewV4()
		key = sarama.ByteEncoder(u1.Bytes())
	} else {
		key = sarama.StringEncoder(message.Key)
	}
	if message.Partition <= 0 {
		message.Partition = int32(-1)
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = now
	}

	return &sarama.ProducerMessage{
		Topic:     message.Topic,
		Key:       key,
		Value:     sarama.StringEncoder(message.Value),
		Metadata:  message.Metadata,
		Offset:    message.Offset,
		Partition: message.Partition,
		Timestamp: message.Timestamp,
	}
}
