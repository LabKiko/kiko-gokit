package datalog

import (
	"go.uber.org/zap/zapcore"
)

const (
	filename = "trans"
	basePath = "./datalog"
)

var (
	defaultOption = &config{
		loggerOpts: loggerOpts{
			path:          basePath,
			encoderConfig: defaultEncoderConfig,
		},
		saOpts: saOpts{
			projectName: "default",
		},
	}

	defaultEncoderConfig = zapcore.EncoderConfig{
		CallerKey:      "callerKey",
		StacktraceKey:  "stacktraceKey",
		LineEnding:     zapcore.DefaultLineEnding,
		TimeKey:        "",
		LevelKey:       "",
		NameKey:        "Logger",
		MessageKey:     "event",
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
)

type Option interface {
	apply(*config)
}

type config struct {
	appId    string            // 服务名称
	timeout  int               // 超时时间, 单位毫秒
	metadata map[string]string // 全局元数据

	saOpts     // 神策打点配置
	kafkaOpts  // kafka 配置
	loggerOpts // 日志配置
}

type OptionFunc func(*config)

func (fn OptionFunc) apply(cfg *config) {
	fn(cfg)
}

type loggerOpts struct {
	disable       bool                  // 禁用
	path          string                // 日志路径
	encoderConfig zapcore.EncoderConfig // 编码器配置
}

type kafkaOpts struct {
	brokers []string
}

type saOpts struct {
	serviceName      string // 神策注册服务名
	projectName      string // 项目名, 默认: default
	debug            bool   // 是否开启 DEBUG 模式
	batch            bool   // 是否批量提交
	token            string // 认证凭证
	batchBulkSizeMax int    // 批量最大值
}

func WithAppId(appId string) Option {
	return OptionFunc(func(o *config) {
		o.appId = appId
	})
}

func WithTimeout(timeout int) Option {
	return OptionFunc(func(o *config) {
		o.timeout = timeout
	})
}

// WithMetadata set default fields for the logger
func WithMetadata(md map[string]string) Option {
	return OptionFunc(func(o *config) {
		o.metadata = md
	})
}

// WithLogDisable disable logger
func WithLogDisable(disable bool) Option {
	return OptionFunc(func(o *config) {
		o.loggerOpts.disable = disable
	})
}

func WithBrokers(brokers []string) Option {
	return OptionFunc(func(o *config) {
		o.kafkaOpts.brokers = brokers
	})
}

// WithBasePath set base path.
func WithBasePath(path string) Option {
	return OptionFunc(func(o *config) {
		o.loggerOpts.path = path
	})
}

// WithEncoderConfig set logger encoderConfig
func WithEncoderConfig(encoderConfig zapcore.EncoderConfig) Option {
	return OptionFunc(func(o *config) {
		o.loggerOpts.encoderConfig = encoderConfig
	})
}

func WithDebug(debug bool) Option {
	return OptionFunc(func(o *config) {
		o.saOpts.debug = debug
	})
}

func WithServiceName(serviceName string) Option {
	return OptionFunc(func(o *config) {
		o.saOpts.serviceName = serviceName
	})
}

func WithProjectName(projectName string) Option {
	return OptionFunc(func(o *config) {
		o.saOpts.projectName = projectName
	})
}

func WithToken(token string) Option {
	return OptionFunc(func(o *config) {
		o.saOpts.token = token
	})
}

func WithBatch(batch bool) Option {
	return OptionFunc(func(o *config) {
		o.saOpts.batch = batch
	})
}

func WithBatchBulkSizeMax(size int) Option {
	return OptionFunc(func(o *config) {
		o.saOpts.batchBulkSizeMax = size
	})
}
