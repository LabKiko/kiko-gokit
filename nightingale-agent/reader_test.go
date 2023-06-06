package agent_test

import (
	"context"
	"log"
	"testing"
	"time"

	agent "github.com/LabKiko/kiko-gokit/nightingale-agent"
	"github.com/LabKiko/kiko-gokit/nightingale-agent/types"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
)

func TestNewReader(t *testing.T) {
	reader := agent.NewReader(
		agent.WithTags(map[string]string{"tag": "value"}),
		agent.WithInterval(time.Second*15),
		agent.WithBatchSize(2000),
		agent.WithQueueSize(10000),
	)
	assert.NotNil(t, reader)
}

type MockInput struct {
}

func (input *MockInput) Init() error {
	return nil
}

func (input *MockInput) Gather() ([]*types.Sample, error) {
	return []*types.Sample{
		{
			Metric:    "test_metric",
			Value:     10,
			Labels:    model.LabelSet{"tag": "value", "tag2": "value2"},
			Timestamp: time.Now(),
		},
		{
			Metric:    "test_metric_1",
			Value:     20,
			Labels:    model.LabelSet{"tag": "value", "tag2": "value2"},
			Timestamp: time.Now(),
		},
	}, nil
}

func (input *MockInput) GetInterval() time.Duration {
	return time.Millisecond * 10
}

func (input *MockInput) String() string {
	return "mock_input"
}

func (input *MockInput) Prefix() string {
	return ""
}

func (input *MockInput) Close() error {
	return nil
}

type MockOutput struct {
}

func (output *MockOutput) Write(ctx context.Context, items []*types.Sample) error {
	for _, item := range items {
		log.Println(item)
	}
	return nil
}

func TestReader_Init(t *testing.T) {
	reader := agent.NewReader(
		agent.WithInterval(time.Millisecond*10),
		agent.WithOutputInterval(time.Second*20),
		agent.WithBatchSize(2000),
		agent.WithQueueSize(10000),
		agent.WithTags(map[string]string{"tag": "value"}),
	)
	input := &MockInput{}
	output := &MockOutput{}
	err := reader.Init(input, output)
	assert.Nil(t, err)

	err = reader.Start()
	assert.Nil(t, err)

	assert.Equal(t, input.String(), reader.String())

	time.Sleep(time.Second * 2)
	// assert.Nil(t, err)
	err = reader.Stop()
	assert.Nil(t, err)
	err = reader.Stop()
	assert.Nil(t, err)
}
