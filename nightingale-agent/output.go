package agent

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/LabKiko/kiko-gokit/nightingale-agent/types"
	"github.com/prometheus/prometheus/prompb"
)

// Output interface wraps the Write method.
type Output interface {
	Write(ctx context.Context, items []*types.Sample) error
}

type defaultOutput struct {
	opt *OutputOptions
}

// NewOutput returns a new Output.
func NewOutput(remote RemoteWriter, opts ...OutputOption) Output {
	op := &OutputOptions{
		RemoteWriter: remote,
		Timeout:      5 * time.Second,
	}
	for _, opt := range opts {
		opt(op)
	}

	return &defaultOutput{
		opt: op,
	}
}

// Write writes the given items to the output.
func (out *defaultOutput) Write(ctx context.Context, items []*types.Sample) error {
	var (
		cancel context.CancelFunc
	)

	if out.opt.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, out.opt.Timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	count := len(items)
	series := make([]prompb.TimeSeries, count)
	for i := 0; i < count; i++ {
		series[i] = items[i].ConvertTimeSeries()
	}

	if out.opt.Debug {
		out.PrintMetrics(items)
	}

	if err := out.opt.RemoteWriter.Write(ctx, series); err != nil {
		return err
	}

	return nil
}

// PrintMetrics prints the given items to stdout.
func (out *defaultOutput) PrintMetrics(samples []*types.Sample) {
	for i := 0; i < len(samples); i++ {
		var sb strings.Builder
		sb.WriteString(samples[i].Timestamp.Format("2006-01-02 15:04:05"))
		sb.WriteString(" ")
		sb.WriteString(samples[i].Metric)

		arr := make([]string, 0, len(samples[i].Labels))
		for key, val := range samples[i].Labels {
			arr = append(arr, fmt.Sprintf("%s=%v", key, val))
		}

		sort.Strings(arr)

		for _, pair := range arr {
			sb.WriteString(" ")
			sb.WriteString(pair)
		}

		sb.WriteString(" ")
		sb.WriteString(fmt.Sprint(samples[i].Value))

		fmt.Println(sb.String())
	}
}
