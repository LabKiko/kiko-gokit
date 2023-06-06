package agent_test

import (
	"context"
	"log"
	"testing"
	"time"

	agent "github.com/LabKiko/kiko-gokit/nightingale-agent"
	"github.com/LabKiko/kiko-gokit/nightingale-agent/types"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
)

type MockRemoteWriter struct {
}

func (m *MockRemoteWriter) Write(ctx context.Context, items []prompb.TimeSeries) error {
	for _, item := range items {
		log.Println(item)
	}

	return nil
}

func TestNew2(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	output := agent.NewOutput(&MockRemoteWriter{}, agent.WithDebug(true))
	err := output.Write(ctx, []*types.Sample{
		{
			Metric:    "up",
			Value:     1,
			Labels:    model.LabelSet{"a": "b", "region": "us"},
			Timestamp: time.Now(),
		},
		{
			Metric:    "http_request_total",
			Value:     1002,
			Labels:    model.LabelSet{"code": "200"},
			Timestamp: time.Now(),
		},
	})

	assert.Nil(t, err)
}
