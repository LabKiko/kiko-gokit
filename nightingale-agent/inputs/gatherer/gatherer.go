package gatherer

import (
	"fmt"
	"math"
	"time"

	"github.com/LabKiko/kiko-gokit/logger"
	"github.com/LabKiko/kiko-gokit/nightingale-agent/pkg/prom"
	"github.com/LabKiko/kiko-gokit/nightingale-agent/types"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/model"
)

type Gatherer struct {
	opt *Options
}

// New creates a new Gatherer
func New(gatherer prometheus.Gatherer, opt ...Option) *Gatherer {
	op := &Options{
		Gatherer: gatherer,
		Interval: time.Second * 30,
	}
	for _, o := range opt {
		o(op)
	}

	return &Gatherer{
		opt: op,
	}
}

// Init initializes the Gatherer
func (input *Gatherer) Init() error {
	return nil
}

// Gather gathers metrics
func (input *Gatherer) Gather() ([]*types.Sample, error) {
	gather, err := input.opt.Gatherer.Gather()
	if err != nil {
		return nil, err
	}

	metricFamilies := make(map[string]*dto.MetricFamily)
	for _, metric := range gather {
		metricFamilies[metric.GetName()] = metric
	}

	// read metrics
	now := time.Now()
	samples := make([]*types.Sample, 0, len(metricFamilies))
	for metricName, mf := range metricFamilies {
		if input.opt.IgnoreMetrics != nil && input.opt.IgnoreMetrics.Match(metricName) {
			logger.Debugf("ignore metric %s", metricName)
			continue
		}

		for _, metric := range mf.GetMetric() {
			// reading tags
			tags := input.makeLabels(metric)
			switch mf.GetType() {
			case dto.MetricType_SUMMARY: // 摘要（Summary）
				samples = append(samples, input.Summary(metricName, metric, tags)...)
			case dto.MetricType_HISTOGRAM: // 直方图（Histogram）
				samples = append(samples, input.Histogram(metricName, metric, tags)...)
			default:
				samples = append(samples, input.Standard(metricName, metric, tags)...)
			}
		}
	}

	samples = types.SetTimestamps(now, samples)
	return samples, nil
}

// Standard metric
func (input *Gatherer) Standard(metricName string, metric *dto.Metric, tags map[string]string) []*types.Sample {
	fields := getNameAndValue(metric, metricName)
	samples := make([]*types.Sample, 0, len(fields))
	for metric, value := range fields {
		samples = append(samples, types.NewSample(prom.BuildMetric(input.opt.Prefix, metric, ""), model.SampleValue(value), tags))
	}

	return samples
}

// Summary metric
func (input *Gatherer) Summary(metricName string, metric *dto.Metric, tags map[string]string) []*types.Sample {
	samples := make([]*types.Sample, 0, 2+len(metric.GetSummary().GetQuantile()))
	samples = append(samples,
		types.NewSample(prom.BuildMetric(input.opt.Prefix, metricName, "count"), model.SampleValue(metric.GetSummary().GetSampleCount()), tags),
		types.NewSample(prom.BuildMetric(input.opt.Prefix, metricName, "sum"), model.SampleValue(metric.GetSummary().GetSampleSum()), tags),
	)

	for _, quantile := range metric.GetSummary().GetQuantile() {
		samples = append(samples,
			types.NewSample(prom.BuildMetric(input.opt.Prefix, metricName), model.SampleValue(quantile.GetValue()), tags, map[string]string{"quantile": fmt.Sprint(quantile.GetQuantile())}),
		)
	}
	return samples
}

// Histogram metric
func (input *Gatherer) Histogram(metricName string, metric *dto.Metric, tags map[string]string) []*types.Sample {
	samples := make([]*types.Sample, 0, 3)
	samples = append(samples,
		types.NewSample(prom.BuildMetric(input.opt.Prefix, metricName, "count"), model.SampleValue(metric.GetHistogram().GetSampleCount()), tags),
		types.NewSample(prom.BuildMetric(input.opt.Prefix, metricName, "sum"), model.SampleValue(metric.GetHistogram().GetSampleSum()), tags),
		types.NewSample(prom.BuildMetric(input.opt.Prefix, metricName, "bucket"), model.SampleValue(metric.GetHistogram().GetSampleCount()), tags, map[string]string{"le": "+Inf"}),
	)

	for _, bucket := range metric.GetHistogram().GetBucket() {
		le := fmt.Sprint(bucket.GetUpperBound())
		samples = append(samples,
			types.NewSample(prom.BuildMetric(input.opt.Prefix, metricName, "bucket"), model.SampleValue(bucket.GetCumulativeCount()), tags, map[string]string{"le": le}),
		)
	}

	return samples
}

// Get labels from metric
func (input *Gatherer) makeLabels(m *dto.Metric) map[string]string {
	result := map[string]string{}

	for _, lp := range m.Label {
		if input.opt.IgnoreLabelKeys != nil && input.opt.IgnoreLabelKeys.Match(lp.GetName()) {
			logger.Debugf("ignore label %s", lp.GetName())
			continue
		}
		result[lp.GetName()] = lp.GetValue()
	}

	for key, value := range input.opt.Tags {
		result[key] = value
	}

	return result
}

// Get name and value from metric
func getNameAndValue(m *dto.Metric, metricName string) map[string]float64 {
	fields := make(map[string]float64)
	if m.Gauge != nil {
		if !math.IsNaN(m.GetGauge().GetValue()) {
			fields[metricName] = m.GetGauge().GetValue()
		}
	} else if m.Counter != nil {
		if !math.IsNaN(m.GetCounter().GetValue()) {
			fields[metricName] = m.GetCounter().GetValue()
		}
	} else if m.Untyped != nil {
		if !math.IsNaN(m.GetUntyped().GetValue()) {
			fields[metricName] = m.GetUntyped().GetValue()
		}
	}
	return fields
}

// GetInterval returns the interval of the input.
func (input *Gatherer) GetInterval() time.Duration {
	return input.opt.Interval
}

// Prefix returns the prefix of the input.
func (input *Gatherer) Prefix() string {
	return input.opt.Prefix
}

// String returns the string representation of the input.
func (input *Gatherer) String() string {
	return "gatherer"
}

func (input *Gatherer) Close() error {
	logger.Infof("close gatherer")
	return nil
}
