package datalog

import (
	"context"
	"fmt"
	"io"
	"path"
	"sync"

	"github.com/LabKiko/kiko-gokit/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	closed       int32
	opts         *config
	base         *zap.Logger
	_rollingFile io.Writer

	sync.RWMutex
}

func NewLogger(opts *config) (*Logger, error) {
	enc := zapcore.NewJSONEncoder(opts.loggerOpts.encoderConfig)

	basePath := fmt.Sprintf("%s/%s/datalog", opts.loggerOpts.path, opts.appId)
	rollingFile, err := logger.NewRollingFile(path.Join(basePath, filename), logger.HourlyRolling)
	if err != nil {
		return nil, err
	}

	// file sync
	fileWriteSyncer := zapcore.AddSync(rollingFile)

	zapLog := zap.New(zapcore.NewTee(zapcore.NewCore(enc, fileWriteSyncer, zap.DebugLevel)))
	// if opts.metadata != nil {
	//	dst := make([]zap.Field, 0, len(opts.metadata))
	//	for k, v := range opts.metadata {
	//		dst = append(dst, zap.Any(k, v))
	//	}
	//	zapLog = zapLog.With(dst...)
	// }

	return &Logger{
		opts:         opts,
		base:         zapLog,
		_rollingFile: fileWriteSyncer,
	}, nil
}

func (l *Logger) copyFields(fields map[string]interface{}) []zap.Field {
	dst := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		dst = append(dst, zap.Any(k, v))
	}
	return dst
}

func (l *Logger) Write(ctx context.Context, event *Event, metadata Metadata) error {

	l.base.Info(event.Name, l.copyFields(metadata)...)

	return nil
}

func (l *Logger) Flush() error {
	return l.base.Sync()
}

func (l *Logger) Close() error {
	return l.Flush()
}
