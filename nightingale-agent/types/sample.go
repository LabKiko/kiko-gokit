package types

import (
	"strings"
	"time"

	"github.com/LabKiko/kiko-gokit/logger"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
)

var (
	labelReplacer = strings.NewReplacer("-", "_", ".", "_", " ", "_", "/", "_")
)

type Sample struct {
	Metric    string            `json:"metric"`
	Value     model.SampleValue `json:"value"`
	Labels    model.LabelSet    `json:"labels"`
	Timestamp time.Time         `json:"timestamp"`
}

// NewSample creates a new Sample
func NewSample(metric string, value model.SampleValue, labels ...map[string]string) *Sample {
	sample := &Sample{
		Metric: metric,
		Value:  value,
		Labels: model.LabelSet{},
	}

	for i := 0; i < len(labels); i++ {
		for k, v := range labels[i] {
			if v == "-" {
				continue
			}

			key := model.LabelName(k)
			if !key.IsValid() {
				logger.Warnf("invalid label name: %s", k)
				continue
			}

			value := model.LabelValue(v)
			if !value.IsValid() {
				logger.Warnf("invalid label value: %s", v)
				continue
			}
			sample.Labels[key] = value
		}
	}

	return sample
}

func (sample *Sample) SetTimestamp(timestamp time.Time) {
	sample.Timestamp = timestamp
}

// ConvertTimeSeries converts a prometheus TimeSeries to a Sample
func (sample *Sample) ConvertTimeSeries() prompb.TimeSeries {
	pt := prompb.TimeSeries{}

	timestamp := sample.Timestamp.UnixMilli()
	pt.Samples = append(pt.Samples, prompb.Sample{
		Timestamp: timestamp,
		Value:     float64(sample.Value),
	})

	// add label: metric
	pt.Labels = append(pt.Labels, prompb.Label{
		Name:  model.MetricNameLabel,
		Value: sample.Metric,
	})

	// add other labels
	for k, v := range sample.Labels {
		pt.Labels = append(pt.Labels, prompb.Label{
			Name:  labelReplacer.Replace(string(k)),
			Value: string(v),
		})
	}

	return pt
}

func SetTimestamps(now time.Time, samples []*Sample) []*Sample {
	if now.IsZero() {
		now = time.Now()
	}
	for _, sample := range samples {
		sample.SetTimestamp(now)
	}
	return samples
}
