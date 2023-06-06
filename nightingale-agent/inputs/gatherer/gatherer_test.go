package gatherer

import (
	"errors"
	"testing"
	"time"

	"github.com/LabKiko/kiko-gokit/nightingale-agent/pkg/filter"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	compile, err := filter.Compile([]string{"ignore1", "ignore2"})
	assert.Nil(t, err)

	metrics, err := filter.Compile([]string{"ignore1", "ignore2"})
	assert.Nil(t, err)

	New(prometheus.DefaultGatherer,
		WithTags(map[string]string{
			"tag1": "value1",
			"tag2": "value2",
			"tag":  "value",
		}),
		WithPrefix("zeus"),
		WithInterval(time.Second*30),
		WithIgnoreMetrics(metrics),
		WithIgnoreLabelKeys(compile),
	)
}

func TestGatherer_Init(t *testing.T) {
	gatherer := New(prometheus.DefaultGatherer,
		WithTags(map[string]string{
			"tag1": "value1",
			"tag2": "value2",
			"tag":  "value",
		}),
		WithPrefix("zeus"),
		WithInterval(time.Second*30),
	)

	err := gatherer.Init()
	assert.Nil(t, err)
}

func TestGatherer_Gather(t *testing.T) {
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

	IgnoreMetrics, err := filter.Compile([]string{"go_goroutines"})
	assert.Nil(t, err)

	IgnoreLabelKey, err := filter.Compile([]string{"zeus"})
	assert.Nil(t, err)
	gatherer := New(prometheus.DefaultGatherer,
		WithTags(map[string]string{
			"tag1": "value1",
			"tag2": "value2",
			"tag":  "value",
		}),
		WithInterval(time.Second*30),
		WithIgnoreMetrics(IgnoreMetrics),
		WithIgnoreLabelKeys(IgnoreLabelKey),
	)

	err = gatherer.Init()
	assert.Nil(t, err)

	metrics, err := gatherer.Gather()
	assert.Nil(t, err)

	for _, m := range metrics {
		t.Log(m)
	}

	assert.NotNil(t, metrics)
}

func TestGatherer_GetInterval(t *testing.T) {
	gatherer := New(prometheus.DefaultGatherer, WithInterval(time.Second*30))
	assert.Equal(t, time.Second*30, gatherer.GetInterval())
}

func TestGatherer_String(t *testing.T) {
	gatherer := New(prometheus.DefaultGatherer)
	assert.Equal(t, "gatherer", gatherer.String())
}

type GathererMock struct {
}

func (g *GathererMock) Gather() ([]*dto.MetricFamily, error) {
	return nil, errors.New("mock error")
}

func TestGatherer_Close(t *testing.T) {
	gatherer := New(&GathererMock{})
	err := gatherer.Init()
	assert.Nil(t, err)

	_, err = gatherer.Gather()
	assert.Equal(t, errors.New("mock error"), err)

	gatherer = New(prometheus.DefaultGatherer)
	err = gatherer.Close()
	assert.Nil(t, err)
}
