/**
* Author: JeffreyBool
* Date: 2021/11/10
* Time: 18:14
* Software: GoLand
 */

package logger

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const (
	msg = "hello there"
)

var (
	keysAndValues = []interface{}{"age", 23, "order", 100}
)

func TestMain(t *testing.M) {
	InitDefaultLogger()

	defer Sync()
	t.Run()
}

func contextWithSpan(operationName string) (context.Context, trace.Span) {
	tracer := otel.Tracer("gokit")
	ctx, span := tracer.Start(context.Background(), operationName, trace.WithSpanKind(trace.SpanKindInternal))
	return ctx, span
}

func TestCopyFields(t *testing.T) {
	fields := map[string]interface{}{
		"a": 1,
		"b": 2,
	}

	newFields := CopyFields(fields)
	assert.NotEqual(t, fields, newFields)
}

func TestLevelLogger(t *testing.T) {
	Debug(msg)
	Info(msg)
	//SetLevel(DebugLevel)
	Debug(msg)
	Warn(msg)
	Error(msg)
	Sync()
	//Fatal(msg)
}

func TestWithError(t *testing.T) {
	DefaultLogger.WithFields(map[string]interface{}{
		"age":   22,
		"order": 100,
	}).WithError(errors.New("logger error")).Info(msg)
}

func TestDefault_Debug(t *testing.T) {
	SetLevel(DebugLevel)
	Debug(msg)

	SetLevel(InfoLevel)
	Debug(msg)
	Sync()
}

func TestDefault_LevelEnablerFunc(t *testing.T) {
	log := &Default{atomicLevel: zap.NewAtomicLevelAt(zap.DebugLevel)}
	assert.True(t, log.LevelEnablerFunc(zap.DebugLevel)(zap.DebugLevel))
}

func TestDefault_Debugf(t *testing.T) {
	Debugf(msg)
}

func TestDefault_Debugw(t *testing.T) {
	Debugw(msg, keysAndValues...)
}

func TestDefault_Error(t *testing.T) {
	Error(msg)
}

func TestDefault_Errorf(t *testing.T) {
	Errorf(msg)
}

func TestDefault_Errorw(t *testing.T) {
	Errorw(msg, keysAndValues...)
}

func TestDefault_Fatal(t *testing.T) {
	Fatal(msg)
}

func TestDefault_Fatalf(t *testing.T) {
	Fatalf(msg)
}

func TestDefault_Fatalw(t *testing.T) {
	Fatalw(msg, keysAndValues...)
}

func TestDefault_Info(t *testing.T) {
	Info(msg)
}

func TestDefault_Infof(t *testing.T) {
	Infof(msg)
}

func TestDefault_Infow(t *testing.T) {
	Infow(msg, keysAndValues...)
}

func TestDefault_Log(t *testing.T) {
	DefaultLogger.Log(InfoLevel, msg, nil, nil)
}

func TestDefault_StdLog(t *testing.T) {
	std := DefaultLogger.StdLog()
	std.Println(msg)
	std.Printf(msg)
}

func TestDefault_Sync(t *testing.T) {
	n := 100
	for i := 0; i < n; i++ {
		Infof("msg info: [%d]", i)
	}
	Sync()
}

func TestDefault_Warn(t *testing.T) {
	Warn(msg)
}

func TestDefault_Warnf(t *testing.T) {
	Warnf(msg)
}

func TestDefault_Warnw(t *testing.T) {
	Warnw(msg, keysAndValues...)
}

func TestDefault_WithCallDepth(t *testing.T) {
	log := New(WithBasePath("../logs"), WithConsole(true))
	log.WithCallDepth(0).Info(msg)
}

func TestDefault_WithContext(t *testing.T) {
	ctx, span := contextWithSpan("TestDefault_WithContext")
	defer span.End()
	WithContext(ctx).Info(msg)
}

func TestDefault_WithFields(t *testing.T) {
	DefaultLogger.WithFields(map[string]interface{}{
		"age":   22,
		"order": 100,
	}).Info(msg)
}

func TestDefault_clone(t *testing.T) {
	log := New(WithBasePath("../logs"), WithConsole(true))
	assert.Equal(t, log, log)
	assert.NotEqual(t, log, log.clone())
}

func TestDefault_createOutput(t *testing.T) {
	log := New(WithBasePath("../logs"), WithConsole(true))
	writeSyncer, err := log.createOutput(infoFilename)
	if err != nil {
		assert.Error(t, err)
	}

	writeSyncer.Write([]byte(msg))
	writeSyncer.Sync()
}

func TestDefault_log(t *testing.T) {
	log := New(WithBasePath("../logs"), WithConsole(true))
	log.log(DebugLevel, msg, nil, nil)
}

func TestDefault_setUp(t *testing.T) {
	log := New(WithBasePath("../logs"), WithConsole(true))
	if err := log.setUp(); err != nil {
		assert.Error(t, err)
	}
}

func TestDefault_setupWithConsole(t *testing.T) {

}

func TestDefault_setupWithFile(t *testing.T) {

}

func TestDefault_setupWithFiles(t *testing.T) {

}

func TestDefault_sweetenFields(t *testing.T) {

}

func TestDefault_withOptions(t *testing.T) {

}

func TestNewDefaultLogger(t *testing.T) {

}

func Test_getMessage(t *testing.T) {

}

func Test_invalidPair_MarshalLogObject(t *testing.T) {

}

func Test_invalidPairs_MarshalLogArray(t *testing.T) {

}

func Test_logWriter_Write(t *testing.T) {

}

func TestSlowLogger(t *testing.T) {
	slow := New(
		WithBasePath("../logs"),
		WithConsole(true),
		WithDisableDisk(false),
		WithLevel(InfoLevel),
		WithFilename("slow"),
		WithFields(map[string]interface{}{
			"app_id":      "mt",
			"instance_id": "JeffreyBool",
		}),
	)

	defer slow.Sync()
	slow.Debug(msg)
	slow.SetLevel(ErrorLevel)
	slow.Debug(msg)
	slow.SetLevel(InfoLevel)
	slow.Info(msg)
}

func TestFileLogger(t *testing.T) {
	stat := New(WithBasePath("../logs"),
		WithConsole(true),
		WithDisableDisk(false),
		WithFilename("stat"),
		WithFields(map[string]interface{}{
			"app_id":      "mt",
			"instance_id": "JeffreyBool",
		}),
	)
	defer stat.Sync()

	stat.Debug(msg)
}

func TestLogger(t *testing.T) {
	log := New(WithBasePath("../logs"),
		WithConsole(true),
		WithDisableDisk(true),
		WithFields(map[string]interface{}{
			"app_id":      "mt",
			"instance_id": "JeffreyBool",
		}),
	)
	defer log.Sync()

	log.Debug(msg)
}
