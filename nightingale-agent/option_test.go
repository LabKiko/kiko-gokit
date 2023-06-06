package agent_test

import (
	"testing"
	"time"

	agent "github.com/LabKiko/kiko-gokit/nightingale-agent"
	"github.com/stretchr/testify/assert"
)

func TestWithTags(t *testing.T) {
	tags := map[string]string{"tag": "value"}
	options := &agent.ReaderOptions{}
	agent.WithTags(tags)(options)
	assert.Equal(t, tags, options.Tags)
}

func TestWithInterval(t *testing.T) {
	interval := time.Second * 15
	options := &agent.ReaderOptions{}
	agent.WithInterval(interval)(options)
	assert.Equal(t, interval, options.Interval)
}

func TestWithBatchSize(t *testing.T) {
	batchSize := 2000
	options := &agent.ReaderOptions{}
	agent.WithBatchSize(batchSize)(options)
	assert.Equal(t, batchSize, options.BatchSize)
}

func TestWithQueueSize(t *testing.T) {
	queueSize := 10000
	options := &agent.ReaderOptions{}
	agent.WithQueueSize(queueSize)(options)
	assert.Equal(t, queueSize, options.QueueSize)
}

func TestWithBasicPassword(t *testing.T) {
	password := "password"
	options := &agent.RemoteWriterOptions{}
	agent.WithBasicPassword(password)(options)
	assert.Equal(t, password, options.BasicPassword)
}

func TestWithBasicUsername(t *testing.T) {
	username := "username"
	options := &agent.RemoteWriterOptions{}
	agent.WithBasicUsername(username)(options)
	assert.Equal(t, username, options.BasicUsername)
}

func TestWithDebug(t *testing.T) {
	options := &agent.OutputOptions{}
	agent.WithDebug(true)(options)
	assert.True(t, options.Debug)
}

func TestWithOutputInterval(t *testing.T) {
	interval := time.Second * 15
	options := &agent.ReaderOptions{}
	agent.WithOutputInterval(interval)(options)
	assert.Equal(t, interval, options.OutputInterval)
}

func TestWithDialTimeout(t *testing.T) {
	timeout := time.Second * 15
	options := &agent.RemoteWriterOptions{}
	agent.WithDialTimeout(timeout)(options)
	assert.Equal(t, timeout, options.DialTimeout)
}

func TestWithMaxIdleConnsPerHost(t *testing.T) {
	v := 2000
	options := &agent.RemoteWriterOptions{}
	agent.WithMaxIdleConnsPerHost(v)(options)
	assert.Equal(t, v, options.MaxIdleConnsPerHost)
}

func TestWithResponseHeaderTimeout(t *testing.T) {
	timeout := time.Second * 15
	options := &agent.RemoteWriterOptions{}
	agent.WithResponseHeaderTimeout(timeout)(options)
	assert.Equal(t, timeout, options.ResponseHeaderTimeout)
}

func TestWithTimeout(t *testing.T) {
	timeout := time.Second * 15
	options := &agent.OutputOptions{}
	agent.WithTimeout(timeout)(options)
	assert.Equal(t, timeout, options.Timeout)
}

func TestWithInput(t *testing.T) {
	input := &MockInput{}
	options := &agent.Options{}
	agent.WithInput(input)(options)
	assert.Equal(t, input, options.Inputs[0])
}

func TestWithOutput(t *testing.T) {
	output := &MockOutput{}
	options := &agent.Options{}
	agent.WithOutput(output)(options)
	assert.Equal(t, output, options.Output)
}
