package datalog

import (
	"time"
)

type DistinctType int8

const (
	Guest DistinctType = iota // 访客
	User                      // 用户
)

// IsGuest 是否游客
func (d DistinctType) IsGuest() bool {
	return d == Guest
}

// IsUser 是否用户
func (d DistinctType) IsUser() bool {
	return d == User
}

type Event struct {
	Name         string       // 事件名称
	DistinctId   string       // 唯一标识 ID; 用户UID, 或者游客ID
	DistinctType DistinctType // 唯一标识类型：（0-访客，1-用户）
	Time         time.Time    // 事件时间
}

type Metadata map[string]interface{}
