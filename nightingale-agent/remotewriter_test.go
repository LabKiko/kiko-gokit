package agent_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	agent "github.com/LabKiko/kiko-gokit/nightingale-agent"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
)

// Serve any http request with a response of N KB of spaces.
type serveSpaces struct {
	sizeKB int
}

func (t serveSpaces) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	kb := bytes.Repeat([]byte{' '}, 1024)
	for i := 0; i < t.sizeKB; i++ {
		w.Write(kb)
	}
}

func TestNewRemoteWriter(t *testing.T) {
	writer, err := agent.NewRemoteWriter("http://localhost:8080/prometheus/api/v1/query?query=up",
		agent.WithBasicUsername("user"),
		agent.WithBasicPassword("pass"),
		agent.WithResponseHeaderTimeout(time.Second),
		agent.WithDialTimeout(time.Second),
		agent.WithMaxIdleConnsPerHost(100),
	)
	assert.Nil(t, err)
	assert.NotNil(t, writer)
}

func TestRemoteWriter_Writer(t *testing.T) {
	testServer := httptest.NewServer(serveSpaces{2000})
	defer testServer.Close()

	endpoint := testServer.URL + "/prometheus/api/v1/query?query=up"
	t.Log(endpoint)
	writer, err := agent.NewRemoteWriter(endpoint)
	assert.Nil(t, err)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	items := []prompb.TimeSeries{
		{Labels: []prompb.Label{{Name: "a", Value: "b"}}},
		{Labels: []prompb.Label{{Name: "a", Value: "b3"}, {Name: "region", Value: "us"}}},
		{Labels: []prompb.Label{{Name: "a", Value: "b2"}, {Name: "region", Value: "europe"}}},
	}
	err = writer.Write(ctx, items)
	assert.Nil(t, err)
}
