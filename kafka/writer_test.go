/**
* Author: JeffreyBool
* Date: 2021/11/6
* Time: 17:58
* Software: GoLand
 */

package kafka_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/LabKiko/kiko-gokit/kafka"
	"github.com/LabKiko/kiko-gokit/trace"
)

type Message struct {
	Index int64     `json:"id"`
	T     time.Time `json:"t"`
}

func TestWriter_SendMessage(t *testing.T) {
	_, err := trace.New(trace.WithName("kafka-writer"))
	if err != nil {
		t.Fatal(err)
	}

	writer, err := kafka.NewWriter([]string{"10.130.12.10:9092"})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		msg := Message{
			Index: time.Now().Unix(),
			T:     time.Now(),
		}
		err = writer.SendMessage(context.Background(), &kafka.ProducerMessage{
			Topic: "infra.datalog.biz-test.data",
			Value: []byte(`{"ts":"2021-12-10T16:17:40.327+0800","caller":"datalog/logx_test.go:36","topic":"infra.datalog.biz-test.data","app_id":"infra.datalog.biz","instance_id":"JeffreyBool","msg":hahah} `),
		})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("消息编号:%d, 发送消息成功时间: %s\n", msg.Index, msg.T.Format("2006-01-02 15:04:05.999"))
	}

	time.Sleep(time.Second * 6)
}
