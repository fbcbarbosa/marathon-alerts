package main

import (
	"fmt"
	"time"

	"github.com/gambol99/go-marathon"
)

type Checker interface {
	Name() string
	Check(marathon.Application) AppCheck
}

// Checks for minimum instances of an app running with respect to total # of instances that is
// supposed to run
type MinHealthyTasks struct {
	// DefaultWarningThreshold - overriden using alerts.min-instances.warning
	DefaultWarningThreshold float32
	// DefaultFailThreshold - overriden using alerts.min-instances.fail
	DefaultFailThreshold float32
}

func (n *MinHealthyTasks) Name() string {
	return "min-healthy"
}

func (n *MinHealthyTasks) Check(app marathon.Application) AppCheck {
	result := Pass
	currentlyRunning := float32(app.TasksHealthy)
	message := fmt.Sprintf("Only %d are healthy out of total %d", int(currentlyRunning), app.Instances)
	// fmt.Printf("%s has %f healthy instances running out of %d\n", app.ID, currentlyRunning, app.Instances)

	if currentlyRunning == 0.0 && app.Instances > 0 {
		result = Fail
	} else if currentlyRunning > 0.0 && currentlyRunning < n.DefaultFailThreshold*float32(app.Instances) {
		result = Fail
	} else if currentlyRunning < n.DefaultWarningThreshold*float32(app.Instances) {
		result = Warning
	} else {
		message = fmt.Sprintf("We now have %d healthy out of total %d", app.TasksHealthy, app.Instances)
	}

	return AppCheck{
		App:       app.ID,
		Labels:    app.Labels,
		CheckName: n.Name(),
		Result:    result,
		Message:   message,
		Timestamp: time.Now(),
	}
}
