package datalog

import (
	"context"
	"testing"
	"time"

	"github.com/LabKiko/kiko-gokit/logger"
)

func TestNewLogger(t *testing.T) {
	opts := &config{
		appId:   "infra.bff.feeds",
		timeout: 100,
		metadata: map[string]string{
			"instance_id": "JeffreyBool",
		},
		loggerOpts: loggerOpts{
			path:          "../data",
			encoderConfig: defaultOption.loggerOpts.encoderConfig,
		},
	}

	NewLogger(opts)
}

func TestLogger_Write(t *testing.T) {
	opts := &config{
		appId:   "infra.bff.feeds",
		timeout: 100,
		metadata: map[string]string{
			"app_id": "JeffreyBool",
		},
		loggerOpts: loggerOpts{
			path:          defaultOption.loggerOpts.path,
			encoderConfig: defaultOption.loggerOpts.encoderConfig,
		},
	}

	logger, err := NewLogger(opts)
	if err != nil {
		t.Fatal(err)
	}

	err = logger.Write(context.Background(), &Event{
		Name:         "datalog.test",
		DistinctId:   "10000",
		DistinctType: User,
		Time:         time.Now(),
	}, map[string]interface{}{
		"order_id": 100,
		"age":      25,
	})
	if err != nil {
		return
	}

	if err != nil {
		t.Fatal(err)
	}

	logger.Flush()
}

func TestFun(t *testing.T) {
	logger.InitDefaultLogger()

	TestLogger_Write(t)
}
