package agent_test

import (
	"net/http/httptest"
	"testing"
	"time"

	agent "github.com/LabKiko/kiko-gokit/nightingale-agent"
	"github.com/LabKiko/kiko-gokit/nightingale-agent/inputs/gatherer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestNewAgent(t *testing.T) {
	testServer := httptest.NewServer(serveSpaces{2000})
	defer testServer.Close()

	endpoint := testServer.URL + "/prometheus/api/v1/query?query=up"
	a, err := agent.New(endpoint)
	assert.Nil(t, err)

	assert.NotNil(t, a)
}

func TestAgent_Start(t *testing.T) {
	testServer := httptest.NewServer(serveSpaces{2000})
	defer testServer.Close()

	input := &MockInput{}
	output := &MockOutput{}
	endpoint := testServer.URL + "/prometheus/api/v1/query?query=up"
	a, err := agent.New(endpoint,
		agent.WithInput(input),
		agent.WithOutput(output),
		agent.WithReaderInterval(time.Millisecond*10),
		agent.WithReaderOutputInterval(time.Millisecond*20),
		agent.WithReaderBatchSize(2000),
		agent.WithReaderQueueSize(10000),
		agent.WithReaderTags(map[string]string{"tag": "value"}),
	)
	assert.Nil(t, err)

	err = a.Start()
	assert.Nil(t, err)

	time.Sleep(time.Second * 2)
	err = a.Stop()
	assert.Nil(t, err)
}

func TestAgentData(t *testing.T) {
	his := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "test_histogram",
			Help: "helpless",
		},
		[]string{"tag1", "tag2"},
	)
	his.WithLabelValues("value1", "value2").Observe(1.0)
	his.WithLabelValues("value1", "value2").Observe(3.23)
	his.WithLabelValues("value1", "value2").Observe(78.99)
	his.WithLabelValues("value1", "value2").Observe(6.999)

	summaryVec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "test_summary",
		Help: "helpless",
	}, []string{"tag1", "tag2"})
	summaryVec.WithLabelValues("value1", "value2").Observe(455.0)
	summaryVec.WithLabelValues("value1", "value2").Observe(54.23)
	summaryVec.WithLabelValues("value1", "value2").Observe(380.99)

	counterVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "test_counter",
		Help: "helpless",
	}, []string{"tag1", "tag2"})
	counterVec.WithLabelValues("value1", "value2").Add(10)
	counterVec.WithLabelValues("value1", "value2").Add(25)

	gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "test_gauge",
		Help: "helpless",
	}, []string{"tag1", "tag2", "zeus"})
	gaugeVec.WithLabelValues("value1", "value2", "1111").Set(10)
	gaugeVec.WithLabelValues("value1", "value2", "222").Set(25)
	gaugeVec.WithLabelValues("value1", "value2", "3333").SetToCurrentTime()

	prometheus.MustRegister(his, summaryVec, counterVec, gaugeVec)

	input := gatherer.New(prometheus.DefaultGatherer, gatherer.WithInterval(time.Millisecond*10))
	output := &MockOutput{}
	a, err := agent.New("endpoint",
		agent.WithInput(input),
		agent.WithOutput(output),
		agent.WithReaderInterval(time.Millisecond*10),
		agent.WithReaderOutputInterval(time.Second*20),
		agent.WithReaderBatchSize(2000),
		agent.WithReaderQueueSize(10000),
	)
	assert.Nil(t, err)
	err = a.Start()
	assert.Nil(t, err)

	time.Sleep(time.Second * 2)
	err = a.Stop()
	assert.Nil(t, err)
}
