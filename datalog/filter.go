package datalog

import (
	"context"
)

// FilterOption is filter option.
type FilterOption func(*Filter)

// FilterKey with filter key.
func FilterKey(key ...string) FilterOption {
	return func(o *Filter) {
		for _, v := range key {
			o.key[v] = struct{}{}
		}
	}
}

// FilterValue with filter value.
func FilterValue(value ...string) FilterOption {
	return func(o *Filter) {
		for _, v := range value {
			o.value[v] = struct{}{}
		}
	}
}

// FilterFunc with filter func.
func FilterFunc(f func(ctx context.Context, event *Event, metadata Metadata) bool) FilterOption {
	return func(o *Filter) {
		o.filter = f
	}
}

// Filter is a logger filter.
type Filter struct {
	exporter Exporter
	key      map[interface{}]struct{}
	value    map[interface{}]struct{}
	filter   func(ctx context.Context, event *Event, metadata Metadata) bool
}

// NewFilter new a logger filter.
func NewFilter(exporter Exporter, opts ...FilterOption) *Filter {
	options := Filter{
		exporter: exporter,
		key:      make(map[interface{}]struct{}),
		value:    make(map[interface{}]struct{}),
	}
	for _, o := range opts {
		o(&options)
	}
	return &options
}

func (f *Filter) Write(ctx context.Context, event *Event, metadata Metadata) error {
	if f.filter != nil && f.filter(ctx, event, metadata) {
		return nil
	}
	if len(f.key) > 0 || len(f.value) > 0 {
		for k, _ := range metadata {
			if _, ok := f.key[k]; ok {
				delete(metadata, k)
				continue
			}
		}
	}

	return f.exporter.Write(ctx, event, metadata)
}

func (f *Filter) Flush() error {
	return f.exporter.Flush()
}

func (f *Filter) Close() error {
	return f.exporter.Close()
}
