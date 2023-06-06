package datalog

import (
	"context"
	"testing"
	"time"

	"github.com/LabKiko/kiko-gokit/datalog/attribute"
	"github.com/LabKiko/kiko-gokit/logger"
	"github.com/stretchr/testify/assert"
)

var datalog DataLog

func TestMain(m *testing.M) {
	var (
		err error
	)

	datalog, err = Dial("infra.bff.feeds",
		WithBrokers([]string{"10.130.12.10:9092"}),
		WithBasePath("../data"),
		WithServiceName("fenmiaozhen"),
		WithProjectName(""),
		WithToken("e4f744a7b594fbc1"),
		WithDebug(true),
	)
	if err != nil {
		panic(err)
	}

	defer datalog.Close()

	logger.InitDefaultLogger()

	m.Run()
}

func TestDatalogProvider_Write(t *testing.T) {
	attributes := make([]attribute.KeyValue, 0)
	attributes = append(attributes,
		attribute.String("biz_system", "sona"),
		attribute.String("team_id", "1502564143662628864"),
		attribute.String("department_id", "1503300425254699008"),
	)

	err := datalog.Write(context.Background(), &Event{
		Name:         "adddepartmentsuccess",
		DistinctId:   "1387639968637124617",
		DistinctType: User,
		Time:         time.Now(),
	}, attributes...)

	assert.Nil(t, err)
}
