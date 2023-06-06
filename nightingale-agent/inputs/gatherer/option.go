package gatherer

import (
	"time"

	"github.com/LabKiko/kiko-gokit/nightingale-agent/pkg/filter"
	"github.com/prometheus/client_golang/prometheus"
)

type Option func(o *Options)

type Options struct {
	Gatherer        prometheus.Gatherer
	Prefix          string
	IgnoreMetrics   filter.Filter
	IgnoreLabelKeys filter.Filter
	Tags            map[string]string
	Interval        time.Duration
}

// WithPrefix sets the prefix of the gatherer
func WithPrefix(prefix string) Option {
	return func(o *Options) {
		o.Prefix = prefix
	}
}

// WithIgnoreMetrics sets the metrics of the gatherer
func WithIgnoreMetrics(ignoreMetrics filter.Filter) Option {
	return func(o *Options) {
		o.IgnoreMetrics = ignoreMetrics
	}
}

// WithIgnoreLabelKeys sets the label keys of the gatherer
func WithIgnoreLabelKeys(ignoreLabelKeys filter.Filter) Option {
	return func(o *Options) {
		o.IgnoreLabelKeys = ignoreLabelKeys
	}
}

// WithTags sets the tags of the gatherer
func WithTags(tags map[string]string) Option {
	return func(o *Options) {
		o.Tags = tags
	}
}

// WithInterval sets the interval of the gatherer
func WithInterval(interval time.Duration) Option {
	return func(o *Options) {
		o.Interval = interval
	}
}
