package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	std "log"
	"os"
	"path"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	debugFilename = "debug"
	infoFilename  = "info"
	warnFilename  = "warn"
	errorFilename = "error"
	fatalFilename = "fatal"
)

var _ Logger = (*Default)(nil)

// Default default logger achieve
type Default struct {
	opt           Option
	base          *zap.Logger
	_rollingFiles []io.Writer
	atomicLevel   zap.AtomicLevel
}

func New(opts ...Options) *Default {
	log := &Default{
		atomicLevel: zap.NewAtomicLevelAt(zap.InfoLevel),
	}
	log.withOptions(opts...)
	if err := log.setUp(); err != nil {
		Must(err)
	}

	return log
}

// Must checks if err is nil, otherwise logs the err and exits.
func Must(err error) {
	if err != nil {
		std.Fatal(err)
	}
}

func (l *Default) withOptions(opts ...Options) {
	opt := defaultOption
	for _, o := range opts {
		o(&opt)
	}

	l.atomicLevel = zap.NewAtomicLevelAt(opt.level.unmarshalZapLevel())
	l.opt = opt
}

func (l *Default) setUp() error {
	var (
		core []zapcore.Core
	)

	if l.opt.filename != "" && l.opt.console { // 指定文件终端输出
		cores, err := l.setupWithFileConsole()
		if err != nil {
			return err
		}
		core = append(core, cores...)
	} else if l.opt.filename == "" && l.opt.console { // 开启终端输出
		cores, err := l.setupWithConsole()
		if err != nil {
			return err
		}
		core = append(core, cores...)
	}

	// 指定文件日志
	if l.opt.filename != "" && !l.opt.disableDisk {
		cores, err := l.setupWithFile()
		if err != nil {
			return err
		}
		core = append(core, cores...)
	} else if !l.opt.disableDisk && l.opt.filename == "" {
		cores, err := l.setupWithFiles()
		if err != nil {
			return err
		}
		core = append(core, cores...)
	}

	zapLog := zap.New(zapcore.NewTee(core...)).WithOptions(zap.AddCaller(), zap.AddCallerSkip(l.opt.callerSkip))
	if l.opt.fields != nil {
		keyVals := CopyFields(l.opt.fields)
		zapLog = zapLog.With(keyVals...)
	}
	if l.opt.namespace != "" {
		zapLog.With(zap.Namespace(l.opt.namespace))
	}

	l.base = zapLog

	return nil
}

func (l *Default) setupWithFile() ([]zapcore.Core, error) {
	if err := l.Sync(); err != nil {
		return nil, err
	}

	cores := make([]zapcore.Core, 0, 1)
	var enc zapcore.Encoder
	if l.opt.encoder.IsConsole() {
		enc = zapcore.NewConsoleEncoder(l.opt.encoderConfig)
	} else {
		enc = zapcore.NewJSONEncoder(l.opt.encoderConfig)
	}

	syncerRolling, err := l.createOutput(l.opt.filename)

	if err != nil {
		return nil, err
	}

	cores = append(cores,
		zapcore.NewCore(enc, syncerRolling, l.atomicLevel),
	)

	l._rollingFiles = append(l._rollingFiles, []io.Writer{syncerRolling}...)

	return cores, nil
}

func (l *Default) LevelEnablerFunc(level zapcore.Level) LevelEnablerFunc {
	return func(lvl zapcore.Level) bool {
		if level == zapcore.FatalLevel {
			return l.atomicLevel.Level() <= level && lvl >= level
		}
		return l.atomicLevel.Level() <= level && lvl == level
	}
}

func (l *Default) setupWithFiles() ([]zapcore.Core, error) {
	var (
		err   error
		cores = make([]zapcore.Core, 0, 5)
		syncerRollingDebug, syncerRollingInfo, syncerRollingWarn,
		syncerRollingError, syncerRollingFatal zapcore.WriteSyncer
	)

	var enc zapcore.Encoder
	if l.opt.encoder.IsConsole() {
		enc = zapcore.NewConsoleEncoder(l.opt.encoderConfig)
	} else {
		enc = zapcore.NewJSONEncoder(l.opt.encoderConfig)
	}

	if err := l.Sync(); err != nil {
		return nil, err
	}

	syncerRollingDebug, err = l.createOutput(debugFilename)
	if err != nil {
		return nil, err
	}
	syncerRollingInfo, err = l.createOutput(infoFilename)
	if err != nil {
		return nil, err
	}
	syncerRollingWarn, err = l.createOutput(warnFilename)
	if err != nil {
		return nil, err
	}
	syncerRollingError, err = l.createOutput(errorFilename)
	if err != nil {
		return nil, err
	}
	syncerRollingFatal, err = l.createOutput(fatalFilename)
	if err != nil {
		return nil, err
	}

	cores = append(cores,
		zapcore.NewCore(enc, syncerRollingDebug, l.LevelEnablerFunc(zap.DebugLevel)),
		zapcore.NewCore(enc, syncerRollingInfo, l.LevelEnablerFunc(zap.InfoLevel)),
		zapcore.NewCore(enc, syncerRollingWarn, l.LevelEnablerFunc(zap.WarnLevel)),
		zapcore.NewCore(enc, syncerRollingError, l.LevelEnablerFunc(zap.ErrorLevel)),
		zapcore.NewCore(enc, syncerRollingFatal, l.LevelEnablerFunc(zap.FatalLevel)),
	)

	l._rollingFiles = append(l._rollingFiles, []io.Writer{syncerRollingDebug, syncerRollingInfo, syncerRollingWarn, syncerRollingError, syncerRollingFatal}...)

	return cores, nil
}

func (l *Default) setupWithConsole() ([]zapcore.Core, error) {
	if err := l.Sync(); err != nil {
		return nil, err
	}

	cores := make([]zapcore.Core, 0, 5)

	syncerStdout := zapcore.AddSync(os.Stdout)
	syncerStderr := zapcore.AddSync(os.Stderr)
	var enc zapcore.Encoder
	if l.opt.encoder.IsConsole() {
		enc = zapcore.NewConsoleEncoder(l.opt.encoderConfig)
	} else {
		enc = zapcore.NewJSONEncoder(l.opt.encoderConfig)
	}

	cores = append(cores,
		zapcore.NewCore(enc, syncerStdout, l.LevelEnablerFunc(zap.DebugLevel)),
		zapcore.NewCore(enc, syncerStdout, l.LevelEnablerFunc(zap.InfoLevel)),
		zapcore.NewCore(enc, syncerStdout, l.LevelEnablerFunc(zap.WarnLevel)),
		zapcore.NewCore(enc, syncerStderr, l.LevelEnablerFunc(zap.ErrorLevel)),
		zapcore.NewCore(enc, syncerStderr, l.LevelEnablerFunc(zap.FatalLevel)),
	)

	return cores, nil
}

func (l *Default) setupWithFileConsole() ([]zapcore.Core, error) {
	if err := l.Sync(); err != nil {
		return nil, err
	}

	cores := make([]zapcore.Core, 0, 1)

	syncerStdout := zapcore.AddSync(os.Stdout)
	var enc zapcore.Encoder
	if l.opt.encoder.IsConsole() {
		enc = zapcore.NewConsoleEncoder(l.opt.encoderConfig)
	} else {
		enc = zapcore.NewJSONEncoder(l.opt.encoderConfig)
	}

	cores = append(cores,
		zapcore.NewCore(enc, syncerStdout, l.atomicLevel),
	)

	return cores, nil
}

func (l *Default) createOutput(filename string) (zapcore.WriteSyncer, error) {
	if len(filename) == 0 {
		return nil, ErrLogPathNotSet
	}

	rollingFile, err := NewRollingFile(path.Join(l.opt.basePath, filename), HourlyRolling)
	if err != nil {
		return nil, err
	}

	writeSyncer := zapcore.AddSync(rollingFile)

	return writeSyncer, nil
}

func (l *Default) clone() *Default {
	_copy := *l
	return &_copy
}

func CopyFields(fields map[string]interface{}) []zap.Field {
	dst := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		dst = append(dst, zap.Any(k, v))
	}
	return dst
}

func (l *Default) SetLevel(lv Level) {
	l.opt.level = lv
	l.atomicLevel.SetLevel(lv.unmarshalZapLevel())
}

func (l *Default) extractSpanId(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasSpanID() {
		return spanCtx.SpanID().String()
	}

	return ""
}

func (l *Default) extractTraceId(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}

	return ""
}

// WithContext with context
func (l *Default) WithContext(ctx context.Context) Logger {
	spanId := l.extractSpanId(ctx)
	traceId := l.extractTraceId(ctx)

	fields := map[string]interface{}{}
	if len(spanId) > 0 {
		fields[spanKey] = spanId
	}
	if len(traceId) > 0 {
		fields[traceKey] = traceId
	}

	logger := &Default{
		opt:         l.opt,
		atomicLevel: l.atomicLevel,
		base:        l.base.With(CopyFields(fields)...).WithOptions(zap.AddCallerSkip(0)),
	}
	return logger
}

// WithFields set fields to always be logged
func (l *Default) WithFields(fields map[string]interface{}) Logger {
	logger := &Default{
		opt:         l.opt,
		atomicLevel: l.atomicLevel,
		base:        l.base.With(CopyFields(fields)...).WithOptions(zap.AddCallerSkip(0)),
	}
	return logger
}

// WithCallDepth  with logger call depth.
func (l *Default) WithCallDepth(callDepth int) Logger {
	logger := &Default{
		opt:         l.opt,
		atomicLevel: l.atomicLevel,
		base:        l.base.WithOptions(zap.AddCallerSkip(callDepth)),
	}
	return logger
}

// WithError with logger error
func (l *Default) WithError(err error) Logger {
	logger := &Default{
		opt:         l.opt,
		atomicLevel: l.atomicLevel,
		base: l.base.With(CopyFields(map[string]interface{}{
			"error": err,
		})...).WithOptions(zap.AddCallerSkip(0)),
	}
	return logger
}

// Options logger option value.
func (l *Default) Options() Option {
	return l.opt
}

// getMessage format with Sprint, Sprintf, or neither.
func getMessage(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	return fmt.Sprint(fmtArgs...)
}

type invalidPair struct {
	position   int
	key, value interface{}
}

func (p invalidPair) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("position", int64(p.position))
	zap.Any("key", p.key).AddTo(enc)
	zap.Any("value", p.value).AddTo(enc)
	return nil
}

type invalidPairs []invalidPair

func (ps invalidPairs) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	var err error
	for i := range ps {
		err = multierr.Append(err, enc.AppendObject(ps[i]))
	}
	return err
}

const (
	_oddNumberErrMsg    = "Ignored key without a value."
	_nonStringKeyErrMsg = "Ignored key-value pairs with non-string keys."
)

func (l *Default) sweetenFields(args []interface{}) []zap.Field {
	if len(args) == 0 {
		return nil
	}

	// Allocate enough space for the worst case; if users pass only structured
	// fields, we shouldn't penalize them with extra allocations.
	fields := make([]zap.Field, 0, len(args))
	var invalid invalidPairs

	for i := 0; i < len(args); {
		// This is a strongly-typed field. Consume it and move on.
		if f, ok := args[i].(zap.Field); ok {
			fields = append(fields, f)
			i++
			continue
		}

		// Make sure this element isn't a dangling key.
		if i == len(args)-1 {
			l.base.Error(_oddNumberErrMsg, zap.Any("ignored", args[i]))
			break
		}

		// Consume this value and the next, treating them as a key-value pair. If the
		// key isn't a string, add this pair to the slice of invalid pairs.
		key, val := args[i], args[i+1]
		if keyStr, ok := key.(string); !ok {
			// Subsequent errors are likely, so allocate once up front.
			if cap(invalid) == 0 {
				invalid = make(invalidPairs, 0, len(args)/2)
			}
			invalid = append(invalid, invalidPair{i, key, val})
		} else {
			fields = append(fields, zap.Any(keyStr, val))
		}
		i += 2
	}

	// If we encountered any invalid key-value pairs, log an error.
	if len(invalid) > 0 {
		l.base.Error(_nonStringKeyErrMsg, zap.Array("invalid", invalid))
	}
	return fields
}

// Debug uses fmt.Sprint to construct and log a message.
func (l *Default) Debug(args ...interface{}) {
	l.log(DebugLevel, "", args, nil)
}

// Info uses fmt.Sprint to construct and log a message.
func (l *Default) Info(args ...interface{}) {
	l.log(InfoLevel, "", args, nil)
}

// Warn uses fmt.Sprint to construct and log a message.
func (l *Default) Warn(args ...interface{}) {
	l.log(WarnLevel, "", args, nil)
}

// Error uses fmt.Sprint to construct and log a message.
func (l *Default) Error(args ...interface{}) {
	l.log(ErrorLevel, "", args, nil)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
// Deprecated: 记录消息后，直接调用 os.Exit(1)，这意味着：
// 在其他 goroutine defer 语句不会被执行；
// 各种 buffers 不会被 flush，包括日志的；
// 临时文件或者目录不会被移除；
// 不要使用 fatal 记录日志，而是向调用者返回错误。如果错误一直持续到 main.main。main.main 那就是在退出之前做处理任何清理操作的正确位置。
func (l *Default) Fatal(args ...interface{}) {
	l.log(FatalLevel, "", args, nil)
}

// Debugf uses fmt.Sprintf to log a templated message.
func (l *Default) Debugf(template string, args ...interface{}) {
	l.log(DebugLevel, template, args, nil)
}

// Infof uses fmt.Sprintf to log a templated message.
func (l *Default) Infof(template string, args ...interface{}) {
	l.log(InfoLevel, template, args, nil)
}

// Warnf uses fmt.Sprintf to log a templated message.
func (l *Default) Warnf(template string, args ...interface{}) {
	l.log(WarnLevel, template, args, nil)
}

// Errorf uses fmt.Sprintf to log a templated message.
func (l *Default) Errorf(template string, args ...interface{}) {
	l.log(ErrorLevel, template, args, nil)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
// Deprecated: 记录消息后，直接调用 os.Exit(1)，这意味着：
// 在其他 goroutine defer 语句不会被执行；
// 各种 buffers 不会被 flush，包括日志的；
// 临时文件或者目录不会被移除；
// 不要使用 fatal 记录日志，而是向调用者返回错误。如果错误一直持续到 main.main。main.main 那就是在退出之前做处理任何清理操作的正确位置。
func (l *Default) Fatalf(template string, args ...interface{}) {
	l.log(FatalLevel, template, args, nil)
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//  s.With(keysAndValues).Debug(msg)
func (l *Default) Debugw(msg string, keysAndValues ...interface{}) {
	l.log(DebugLevel, msg, nil, keysAndValues)
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (l *Default) Infow(msg string, keysAndValues ...interface{}) {
	l.log(InfoLevel, msg, nil, keysAndValues)
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (l *Default) Warnw(msg string, keysAndValues ...interface{}) {
	l.log(WarnLevel, msg, nil, keysAndValues)
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (l *Default) Errorw(msg string, keysAndValues ...interface{}) {
	l.log(ErrorLevel, msg, nil, keysAndValues)
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
// Deprecated: 记录消息后，直接调用 os.Exit(1)，这意味着：
// 在其他 goroutine defer 语句不会被执行；
// 各种 buffers 不会被 flush，包括日志的；
// 临时文件或者目录不会被移除；
// 不要使用 fatal 记录日志，而是向调用者返回错误。如果错误一直持续到 main.main。main.main 那就是在退出之前做处理任何清理操作的正确位置。
func (l *Default) Fatalw(msg string, keysAndValues ...interface{}) {
	l.log(FatalLevel, msg, nil, keysAndValues)
}

func (l *Default) Log(level Level, template string, fmtArgs []interface{}, context []interface{}) {
	l.log(level, template, fmtArgs, context)
}

func (l *Default) StdLog() *std.Logger {
	stdLogger := std.New(logWriter{
		logFunc: func() func(msg string, args ...interface{}) {
			logger := &Default{base: l.base.WithOptions(zap.AddCallerSkip(3))}
			return logger.Infof
		},
	}, "", 0)
	return stdLogger
}

type logWriter struct {
	logFunc func() func(msg string, fields ...interface{})
}

func (l logWriter) Write(p []byte) (int, error) {
	p = bytes.TrimSpace(p)
	if l.logFunc != nil {
		l.logFunc()(string(p))
	}
	return len(p), nil
}

func (l *Default) log(level Level, template string, fmtArgs []interface{}, context []interface{}) {
	// If logging at this level is completely disabled, skip the overhead of
	// string formatting.
	if level < DebugLevel || !l.base.Core().Enabled(level.unmarshalZapLevel()) {
		return
	}

	msg := getMessage(template, fmtArgs)
	if ce := l.base.Check(level.unmarshalZapLevel(), msg); ce != nil {
		ce.Write(l.sweetenFields(context)...)
	}
}

func (l *Default) Sync() error {
	if l.base != nil {
		return l.base.Sync()
	}

	for _, w := range l._rollingFiles {
		r, ok := w.(*RollingFile)
		if ok {
			r.Close()
		}
	}

	return nil
}
