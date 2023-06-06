package datalog

import (
	"context"
	"fmt"

	sa "github.com/sensorsdata/sa-sdk-go"
	"github.com/sensorsdata/sa-sdk-go/consumers"
)

var (
	filterFields = []string{"env", "trace_id", "session_id", "app_id", "instance_id", "event", "distinct_id", "user_agent", "time"}
)

type SensorsData struct {
	opts  *config
	write *sa.SensorsAnalytics
}

func NewSensorsData(opts *config) *SensorsData {
	var (
		consumer consumers.Consumer
	)

	serverUrl := fmt.Sprintf("https://%s.datasink.sensorsdata.cn/sa?project=%s&token=%s", opts.saOpts.serviceName, opts.saOpts.projectName, opts.saOpts.token)

	if opts.saOpts.batch {
		consumer, _ = sa.InitBatchConsumer(serverUrl, opts.saOpts.batchBulkSizeMax, opts.timeout)
	} else if opts.saOpts.debug {
		consumer, _ = sa.InitDebugConsumer(serverUrl, true, opts.timeout)
	} else {
		//  DefaultConsumer 是同步发送数据，因此不要在任何线上的服务中使用此 Consumer
		consumer, _ = sa.InitDefaultConsumer(serverUrl, opts.timeout)
	}

	// 使用 Consumer 来构造 SensorsAnalytics 对象
	sensorsAnalytics := sa.InitSensorsAnalytics(consumer, opts.saOpts.projectName, false)
	sensorsAnalytics.RegisterSuperProperties(map[string]interface{}{
		"platform_type": "server",
	})

	return &SensorsData{
		opts:  opts,
		write: &sensorsAnalytics,
	}
}

func (s *SensorsData) Write(ctx context.Context, event *Event, metadata Metadata) error {
	var (
		isLoginId bool
	)

	if event.DistinctType.IsUser() {
		isLoginId = true
	}

	metadata["$time"] = event.Time.Unix()

	return s.write.Track(event.DistinctId, event.Name, metadata, isLoginId)
}

// Flush 将内存数据写入
func (s *SensorsData) Flush() error {
	s.write.Flush()

	return nil
}

// Close 安全关闭
func (s *SensorsData) Close() error {
	s.write.Close()

	return nil
}
