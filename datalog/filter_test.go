package datalog

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFilterAll(t *testing.T) {
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
	logger, err := NewLogger(opts)
	assert.Nil(t, err)
	defer logger.Close()

	log := NewFilter(logger,
		FilterKey("username"),
		FilterValue("hello"),
		FilterFunc(testFilterFunc),
	)

	err = log.Write(context.Background(), &Event{
		Name:         "datalog.test",
		DistinctId:   "10000",
		DistinctType: User,
		Time:         time.Now(),
	}, map[string]interface{}{
		"username": "gokit",
		"password": 123456,
		"value":    "hello",
	})

	assert.Nil(t, err)
}

func TestFilterKey(t *testing.T) {
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
	logger, err := NewLogger(opts)
	assert.Nil(t, err)
	defer logger.Close()

	log := NewFilter(logger, FilterKey("password"))

	err = log.Write(context.Background(), &Event{
		Name:         "datalog.test",
		DistinctId:   "10000",
		DistinctType: User,
		Time:         time.Now(),
	}, map[string]interface{}{
		"password": 123456,
	})

	assert.Nil(t, err)
}

func TestFilterValue(t *testing.T) {
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
	logger, err := NewLogger(opts)
	assert.Nil(t, err)
	defer logger.Close()

	log := NewFilter(logger, FilterValue("debug"))

	err = log.Write(context.Background(), &Event{
		Name:         "datalog.test",
		DistinctId:   "10000",
		DistinctType: User,
		Time:         time.Now(),
	}, map[string]interface{}{
		"test 1": "debug",
	})

	assert.Nil(t, err)
}

func TestFilterFunc(t *testing.T) {
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
	logger, err := NewLogger(opts)
	assert.Nil(t, err)
	defer logger.Close()

	log := NewFilter(logger, FilterFunc(testFilterFunc))

	err = log.Write(context.Background(), &Event{
		Name:         "datalog.test",
		DistinctId:   "10000",
		DistinctType: User,
		Time:         time.Now(),
	}, map[string]interface{}{
		"password": 123456,
	})

	assert.Nil(t, err)
}

func BenchmarkFilterKey(b *testing.B) {
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

	logger, err := NewLogger(opts)
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	b.StartTimer()
	log := NewFilter(logger, FilterKey("password"))
	for i := 0; i < b.N; i++ {
		err = log.Write(context.Background(), &Event{
			Name:         "datalog.test",
			DistinctId:   "10000",
			DistinctType: User,
			Time:         time.Now(),
		}, map[string]interface{}{
			"password": 123456,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
}

func BenchmarkFilterValue(b *testing.B) {
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

	logger, err := NewLogger(opts)
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	b.StartTimer()
	log := NewFilter(logger, FilterValue("password"))
	for i := 0; i < b.N; i++ {
		err = log.Write(context.Background(), &Event{
			Name:         "datalog.test",
			DistinctId:   "10000",
			DistinctType: User,
			Time:         time.Now(),
		}, map[string]interface{}{
			"name": "password",
		})
		if err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
}

func BenchmarkFilterFunc(b *testing.B) {
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

	logger, err := NewLogger(opts)
	if err != nil {
		b.Fatal(err)
	}
	defer logger.Close()

	b.StartTimer()
	log := NewFilter(logger, FilterFunc(testFilterFunc))
	for i := 0; i < b.N; i++ {
		err = log.Write(context.Background(), &Event{
			Name:         "datalog.test",
			DistinctId:   "10000",
			DistinctType: User,
			Time:         time.Now(),
		}, map[string]interface{}{
			"password": "123456",
		})
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
}

func testFilterFunc(ctx context.Context, event *Event, metadata Metadata) bool {
	for k, _ := range metadata {
		if k == "password" {
			metadata[k] = "*****"
		}
	}
	return false
}
