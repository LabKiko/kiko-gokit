package agent

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/prometheus/prompb"
)

// RemoteWriter interface wraps the Write method.
type RemoteWriter interface {
	Write(ctx context.Context, items []prompb.TimeSeries) error
}

// A RemoteWriter is a writer to write StatReport.
type defaultRemoteWriter struct {
	opt    RemoteWriterOptions
	client api.Client
}

// NewRemoteWriter returns a RemoteWriter.
func NewRemoteWriter(endpoint string, opts ...RemoteWriterOption) (RemoteWriter, error) {
	cfg := RemoteWriterOptions{
		Endpoint:              endpoint,
		ResponseHeaderTimeout: time.Second * 5,
		DialTimeout:           time.Second * 2,
		MaxIdleConnsPerHost:   100,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	cli, err := api.NewClient(api.Config{
		Address: endpoint,
		RoundTripper: &http.Transport{
			// TLSClientConfig: tlsConfig,
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout: cfg.DialTimeout,
			}).DialContext,
			ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
			MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
		},
	})
	if err != nil {
		return nil, err
	}

	return &defaultRemoteWriter{
		opt:    cfg,
		client: cli,
	}, nil
}

// Write writes the given items to the remote writer.
func (r *defaultRemoteWriter) Write(ctx context.Context, items []prompb.TimeSeries) error {
	if len(items) == 0 {
		return nil
	}

	req := &prompb.WriteRequest{
		Timeseries: items,
	}
	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	data = snappy.Encode(nil, data)
	httpReq, err := http.NewRequest(http.MethodPost, r.opt.Endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Add("Content-Encoding", "snappy")
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("User-Agent", "agent/prometheus-remote-writer")
	httpReq.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
	if r.opt.BasicUsername != "" {
		httpReq.SetBasicAuth(r.opt.BasicUsername, r.opt.BasicPassword)
	}

	// FIXME: 增加重试机制
	resp, body, err := r.client.Do(ctx, httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode >= 400 {
		err = fmt.Errorf("push data with remote write request got status code: %v, response body: %s", resp.StatusCode, string(body))
		return err
	}

	return nil
}
