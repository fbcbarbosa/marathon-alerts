package main

import (
	"github.com/gambol99/go-marathon"
	"time"
)

type AppCheck struct {
	App       string
	CheckName string
	Result    CheckStatus
	Message   string
	Timestamp time.Time
	Labels    map[string]string
	Times     int
}

type Checker interface {
	Name() string
	Check(marathon.Application) AppCheck
}
