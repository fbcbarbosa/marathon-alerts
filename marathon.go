package main

import "time"

type AppCheck struct {
	App       string
	CheckName string
	Result    CheckStatus
	Message   string
	Timestamp time.Time
	Labels    map[string]string
}
