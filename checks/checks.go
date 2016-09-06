package checks

import (
	"time"

	"github.com/gambol99/go-marathon"
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

type CheckStatus uint8

const (
	Pass     = CheckStatus(99)
	Resolved = CheckStatus(98)
	Warning  = CheckStatus(2)
	Critical = CheckStatus(1)
)

var CheckLevels = [...]CheckStatus{Warning, Critical}

func CheckStatusToString(result CheckStatus) string {
	value := "Unknown"
	switch result {
	case Pass:
		value = "Passed"
	case Resolved:
		value = "Resolved"
	case Warning:
		value = "Warning"
	case Critical:
		value = "Critical"
	}

	return value
}
