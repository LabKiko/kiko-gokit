package datalog

import (
	"context"
	"sync"
	"time"

	"github.com/LabKiko/kiko-gokit/datalog/attribute"
	"github.com/LabKiko/kiko-gokit/logger"
	"github.com/LabKiko/kiko-gokit/trace"
	"go.uber.org/multierr"
)

// Exporter 导出器
type Exporter interface {
	Write(ctx context.Context, event *Event, metadata Metadata) error
	Flush() error
	Close() error
}

type DataLog interface {
	Write(ctx context.Context, event *Event, attributes ...attribute.KeyValue) error
	Flush() error
	Close() error
}

type datalogProvider struct {
	opts      *config
	exporters []Exporter
}

func Dial(appId string, options ...Option) (DataLog, error) {
	cfg := defaultOption
	cfg.appId = appId

	for _, opt := range options {
		opt.apply(cfg)
	}

	c := &datalogProvider{
		opts:      cfg,
		exporters: make([]Exporter, 0, 3),
	}

	c.init()

	return c, nil
}

func (p *datalogProvider) init() {
	p.loggerExporter()
	p.kafkaExporter()
	p.saExporter()
}

func (p *datalogProvider) loggerExporter() {
	if p.opts.loggerOpts.disable {
		return
	}

	log, err := NewLogger(p.opts)
	if err != nil {
		logger.Fatal(err)
	}

	// 神策默认过滤一些自定义打点字段
	p.exporters = append(p.exporters, NewFilter(log, FilterKey("event")))

	logger.Info("datalog init logger exporter success")
}

func (p *datalogProvider) kafkaExporter() {
	if len(p.opts.kafkaOpts.brokers) == 0 {
		return
	}

	kafka, err := NewKafka(p.opts)
	if err != nil {
		logger.Fatal(err)
	}

	p.exporters = append(p.exporters, kafka)

	logger.Info("datalog init kafka exporter success")
}

func (p *datalogProvider) saExporter() {
	if p.opts.saOpts.serviceName == "" || p.opts.saOpts.token == "" {
		return
	}

	filterKeys := make([]FilterOption, 0, len(filterFields))
	for _, field := range filterFields {
		filterKeys = append(filterKeys, FilterKey(field))
	}

	// 神策默认过滤一些自定义打点字段
	p.exporters = append(p.exporters, NewFilter(NewSensorsData(p.opts), filterKeys...))

	logger.Info("datalog init sensorsdata exporter success")
}

func (p *datalogProvider) Write(ctx context.Context, event *Event, attributes ...attribute.KeyValue) error {
	if event.Time.IsZero() {
		event.Time = time.Now()
	}

	// 默认数据
	metadata := make(map[string]interface{})
	metadata["app_id"] = p.opts.appId
	metadata["distinct_id"] = event.DistinctId
	metadata["event"] = event.Name
	metadata["time"] = event.Time.Format(time.RFC3339)
	if ctx != context.Background() && ctx != context.TODO() {
		metadata["trace_id"] = trace.ExtractTraceId(ctx)
	}

	// 添加元数据
	for k, md := range p.opts.metadata {
		metadata[k] = md
	}

	// 添加属性
	for _, kv := range attributes {
		metadata[(string)(kv.Key)] = kv.Value.Emit()
		// switch kv.Value.Type() {
		// //case attribute.BOOLSLICE, attribute.INT64SLICE, attribute.FLOAT64SLICE, attribute.STRINGSLICE:
		// //	json, _ := json.Marshal(kv.Value.AsInterface())
		// //	a := (string)(json)
		// //	metadata[(string)(kv.Key)] = a
		// default:
		//	metadata[(string)(kv.Key)] = kv.Value.Emit()
		// }
	}

	var err error
	var waitGroup = sync.WaitGroup{}
	waitGroup.Add(len(p.exporters))
	for _, exporter := range p.exporters {
		exp := exporter
		go func() {
			defer func() {
				waitGroup.Done()
				if r := recover(); r != nil {
					logger.Errorf("datalog write recover error: %+v", p)
				}
			}()
			err = multierr.Append(err, exp.Write(ctx, event, DeepCopy(metadata)))
		}()
	}

	waitGroup.Wait()

	return err
}

func (p *datalogProvider) Flush() error {
	var err error
	for _, provider := range p.exporters {
		err = multierr.Append(err, provider.Flush())
	}

	return err
}

func (p *datalogProvider) Close() error {
	var err error
	for _, provider := range p.exporters {
		err = multierr.Append(err, provider.Close())
	}

	return err
}
