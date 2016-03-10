package checks

import (
	"testing"

	"github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
)

// === MinInstances ===
func TestMinInstancesWhenEverythingIsFine(t *testing.T) {
	check := MinInstances{
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
	assert.Equal(t, "min-instances", appCheck.CheckName)
	assert.Equal(t, "/foo", appCheck.App)
	assert.Equal(t, "We now have 100 healthy out of total 100", appCheck.Message)
}

func TestMinInstancesWhenWarningThresholdIsMet(t *testing.T) {
	check := MinInstances{
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
	assert.Equal(t, "min-instances", appCheck.CheckName)
	assert.Equal(t, "/foo", appCheck.App)
	assert.Equal(t, "Only 59 are healthy out of total 100", appCheck.Message)
}

func TestMinInstancesWhenWarningThresholdIsMetButOverridenFromAppLabels(t *testing.T) {
	check := MinInstances{
		DefaultCriticalThreshold: 0.4,
		DefaultWarningThreshold:  0.6,
	}
	appLabels := make(map[string]string)
	appLabels["alerts.min-instances.warn.threshold"] = "0.5"
	app := marathon.Application{
		ID:           "/foo",
		Instances:    100,
		TasksHealthy: 59,
		Labels:       appLabels,
	}

	appCheck := check.Check(app)
	assert.Equal(t, Pass, appCheck.Result)
	assert.Equal(t, "min-instances", appCheck.CheckName)
	assert.Equal(t, "/foo", appCheck.App)
	assert.Equal(t, "We now have 59 healthy out of total 100", appCheck.Message)
}

func TestMinInstancesWhenFailThresholdIsMet(t *testing.T) {
	check := MinInstances{
		DefaultCriticalThreshold: 0.5,
		DefaultWarningThreshold:  0.6,
	}
	app := marathon.Application{
		ID:           "/foo",
		Instances:    100,
		TasksHealthy: 47,
		TasksStaged:  2,
	}

	appCheck := check.Check(app)
	assert.Equal(t, Critical, appCheck.Result)
	assert.Equal(t, "min-instances", appCheck.CheckName)
	assert.Equal(t, "/foo", appCheck.App)
	assert.Equal(t, "Only 49 are healthy out of total 100", appCheck.Message)
}

func TestMinInstancesWhenFailThresholdIsMetButOverridenFromAppLabels(t *testing.T) {
	check := MinInstances{
		DefaultCriticalThreshold: 0.5,
		DefaultWarningThreshold:  0.6,
	}
	appLabels := make(map[string]string)
	appLabels["alerts.min-instances.critical.threshold"] = "0.4"
	app := marathon.Application{
		ID:           "/foo",
		Instances:    100,
		TasksHealthy: 48,
		TasksStaged:  1,
		Labels:       appLabels,
	}

	appCheck := check.Check(app)
	assert.Equal(t, Warning, appCheck.Result)
	assert.Equal(t, "min-instances", appCheck.CheckName)
	assert.Equal(t, "/foo", appCheck.App)
	assert.Equal(t, "Only 49 are healthy out of total 100", appCheck.Message)
}

func TestMinInstancesWhenNoTasksAreRunning(t *testing.T) {
	check := MinInstances{
		DefaultCriticalThreshold: 0.5,
		DefaultWarningThreshold:  0.6,
	}
	app := marathon.Application{
		ID:           "/foo",
		Instances:    1,
		TasksHealthy: 0,
		TasksStaged:  0,
	}

	appCheck := check.Check(app)
	assert.Equal(t, Critical, appCheck.Result)
	assert.Equal(t, "min-instances", appCheck.CheckName)
	assert.Equal(t, "/foo", appCheck.App)
	assert.Equal(t, "Only 0 are healthy out of total 1", appCheck.Message)
}
