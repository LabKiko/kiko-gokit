package logger

import (
	"go.uber.org/zap/zapcore"
)

const (
	spanKey          = "span_id"
	traceKey         = "trace_id"
	callerSkipOffset = 2
)

var (
	defaultOption = Option{
		level:         InfoLevel,
		basePath:      "./logs",
		console:       true,
		disableDisk:   true,
		callerSkip:    callerSkipOffset,
		encoderConfig: defaultEncoderConfig,
		fields:        make(map[string]interface{}),
		encoder:       JsonEncoder,
	}

	defaultEncoderConfig = zapcore.EncoderConfig{
		CallerKey:      "caller",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		TimeKey:        "ts",
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "Logger",
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder, // 日期格式改为"ISO8601"，例如："2020-12-16T19:12:48.771+0800"
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
)

type Encoder string

func (e *Encoder) String() string {
	return string(*e)
}

// IsJson Whether json encoder.
func (e *Encoder) IsJson() bool {
	return e.String() == "json"
}

// IsConsole Whether console encoder.
func (e *Encoder) IsConsole() bool {
	return e.String() == "console"
}

const (
	JsonEncoder    Encoder = "json"
	ConsoleEncoder Encoder = "console"
)

type Option struct {
	// The logging level the Logger should log at. default is `InfoLevel`
	level Level
	// basePath defines base path of log file
	basePath string
	//  Logger file name
	filename string
	// enableConsole display logs to standard output
	console bool
	// disableDisk disable rolling file
	disableDisk bool

	callerSkip int

	namespace string

	fields map[string]interface{}

	encoder Encoder

	encoderConfig zapcore.EncoderConfig
}

// Level Get log level.
func (o Option) Level() Level {
	return o.level
}

type Options func(o *Option)

// WithLevel set base path.
func WithLevel(lv Level) Options {
	return func(o *Option) {
		o.level = lv
	}
}

// WithBasePath set base path.
func WithBasePath(path string) Options {
	return func(o *Option) {
		o.basePath = path
	}
}

// WithFilename Logger filename.
func WithFilename(filename string) Options {
	return func(o *Option) {
		o.filename = filename
	}
}

// WithConsole enable console.
func WithConsole(enableConsole bool) Options {
	return func(o *Option) {
		o.console = enableConsole
	}
}

// WithDisableDisk disable disk.
func WithDisableDisk(disableDisk bool) Options {
	return func(o *Option) {
		o.disableDisk = disableDisk
	}
}

// WithCallerSkip increases the number of callers skipped by caller annotation
// (as enabled by the AddCaller option). When building wrappers around the
// Logger and SugaredLogger, supplying this Option prevents base from always
// reporting the wrapper code as the caller.
func WithCallerSkip(skip int) Options {
	return func(o *Option) {
		o.callerSkip = skip + callerSkipOffset
	}
}

// WithFields set default fields for the logger
func WithFields(fields map[string]interface{}) Options {
	return func(o *Option) {
		o.fields = fields
	}
}

// WithEncoder set logger Encoder
func WithEncoder(encoder Encoder) Options {
	return func(o *Option) {
		o.encoder = encoder
	}
}

// WithEncoderConfig set logger encoderConfig
func WithEncoderConfig(encoderConfig zapcore.EncoderConfig) Options {
	return func(o *Option) {
		o.encoderConfig = encoderConfig
	}
}

// WithNamespace creates a named, isolated scope within the logger's context. All
// subsequent fields will be added to the new namespace.
//
// This helps prevent key collisions when injecting loggers into sub-components
// or third-party libraries.
func WithNamespace(name string) Options {
	return func(o *Option) {
		o.namespace = name
	}
}
