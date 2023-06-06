package datalog

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// https://fenmiaozhen.datasink.sensorsdata.cn/sa?token=e4f744a7b594fbc1&project=default

func TestNewSensorsData(t *testing.T) {
	NewSensorsData(&config{
		appId:   "infra.bff.feeds",
		timeout: 100,
		saOpts: saOpts{
			serviceName: "xxx",
			projectName: "default",
			debug:       true,
			token:       "xxx",
		},
	})
}

func TestSensorsData_Write(t *testing.T) {
	sa := NewSensorsData(&config{
		appId:   "infra.bff.feeds",
		timeout: 500,
		saOpts: saOpts{
			serviceName: "fenmiaozhen",
			projectName: "default",
			debug:       true,
			token:       "e4f744a7b594fbc1",
		},
	})

	err := sa.Write(context.Background(), &Event{
		Name:         "adddepartmentsuccess",
		DistinctId:   "1387639968637124617",
		DistinctType: User,
		Time:         time.Now(),
	}, map[string]interface{}{
		"team_id":       "1502564143662628864",
		"department_id": "1503306178447278080",
	})

	assert.Nil(t, err)
}
