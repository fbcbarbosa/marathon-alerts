package main

import (
	"testing"

	"github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
)

func TestMinHealthyTasksWhenEverythingIsFine(t *testing.T) {
	check := MinHealthyTasks{
		DefaultCriticalThreshold: 0.5,
		DefaultWarningThreshold:  0.6,
	}
	app := marathon.Application{
		ID:           "/foo",
		Instances:    100,
		TasksHealthy: 100,
	}

	appCheck := check.Check(app)
	assert.Equal(t, Pass, appCheck.Result)
	assert.Equal(t, "min-healthy", appCheck.CheckName)
	assert.Equal(t, "/foo", appCheck.App)
	assert.Equal(t, "We now have 100 healthy out of total 100", appCheck.Message)
}

func TestMinHealthyTasksWhenWarningThresholdIsMet(t *testing.T) {
	check := MinHealthyTasks{
		DefaultCriticalThreshold: 0.5,
		DefaultWarningThreshold:  0.6,
	}
	app := marathon.Application{
		ID:           "/foo",
		Instances:    100,
		TasksHealthy: 59,
	}

	appCheck := check.Check(app)
	assert.Equal(t, Warning, appCheck.Result)
	assert.Equal(t, "min-healthy", appCheck.CheckName)
	assert.Equal(t, "/foo", appCheck.App)
	assert.Equal(t, "Only 59 are healthy out of total 100", appCheck.Message)
}

func TestMinHealthyTasksWhenWarningThresholdIsMetButOverridenFromAppLabels(t *testing.T) {
	check := MinHealthyTasks{
		DefaultCriticalThreshold: 0.4,
		DefaultWarningThreshold:  0.6,
	}
	appLabels := make(map[string]string)
	appLabels["alerts.min-healthy.warn.threshold"] = "0.5"
	app := marathon.Application{
		ID:           "/foo",
		Instances:    100,
		TasksHealthy: 59,
		Labels:       appLabels,
	}

	appCheck := check.Check(app)
	assert.Equal(t, Pass, appCheck.Result)
	assert.Equal(t, "min-healthy", appCheck.CheckName)
	assert.Equal(t, "/foo", appCheck.App)
	assert.Equal(t, "We now have 59 healthy out of total 100", appCheck.Message)
}

func TestMinHealthyTasksWhenFailThresholdIsMet(t *testing.T) {
	check := MinHealthyTasks{
		DefaultCriticalThreshold: 0.5,
		DefaultWarningThreshold:  0.6,
	}
	app := marathon.Application{
		ID:           "/foo",
		Instances:    100,
		TasksHealthy: 49,
	}

	appCheck := check.Check(app)
	assert.Equal(t, Critical, appCheck.Result)
	assert.Equal(t, "min-healthy", appCheck.CheckName)
	assert.Equal(t, "/foo", appCheck.App)
	assert.Equal(t, "Only 49 are healthy out of total 100", appCheck.Message)
}

func TestMinHealthyTasksWhenFailThresholdIsMetButOverridenFromAppLabels(t *testing.T) {
	check := MinHealthyTasks{
		DefaultCriticalThreshold: 0.5,
		DefaultWarningThreshold:  0.6,
	}
	appLabels := make(map[string]string)
	appLabels["alerts.min-healthy.critical.threshold"] = "0.4"
	app := marathon.Application{
		ID:           "/foo",
		Instances:    100,
		TasksHealthy: 49,
		Labels:       appLabels,
	}

	appCheck := check.Check(app)
	assert.Equal(t, Warning, appCheck.Result)
	assert.Equal(t, "min-healthy", appCheck.CheckName)
	assert.Equal(t, "/foo", appCheck.App)
	assert.Equal(t, "Only 49 are healthy out of total 100", appCheck.Message)
}

func TestMinHealthyTasksWhenNoTasksAreRunning(t *testing.T) {
	check := MinHealthyTasks{
		DefaultCriticalThreshold: 0.5,
		DefaultWarningThreshold:  0.6,
	}
	app := marathon.Application{
		ID:           "/foo",
		Instances:    1,
		TasksHealthy: 0,
	}

	appCheck := check.Check(app)
	assert.Equal(t, Critical, appCheck.Result)
	assert.Equal(t, "min-healthy", appCheck.CheckName)
	assert.Equal(t, "/foo", appCheck.App)
	assert.Equal(t, "Only 0 are healthy out of total 1", appCheck.Message)
}
