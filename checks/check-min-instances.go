package checks

import (
	"fmt"
	"time"

	maps "github.com/ashwanthkumar/golang-utils/maps"
	"github.com/gambol99/go-marathon"
)

// Checks for minimum instances of an app running with respect to total # of instances that is
// supposed to run
type MinInstances struct {
	// DefaultWarningThreshold - overriden using alerts.min-instances.warning
	DefaultWarningThreshold float32
	// DefaultCriticalThreshold - overriden using alerts.min-instances.fail
	DefaultCriticalThreshold float32
}

func (n *MinInstances) Name() string {
	return "min-instances"
}

func (n *MinInstances) Check(app marathon.Application) AppCheck {
	failThreshold := maps.GetFloat32(app.Labels, "alerts.min-instances.critical.threshold", n.DefaultCriticalThreshold)
	warnThreshold := maps.GetFloat32(app.Labels, "alerts.min-instances.warn.threshold", n.DefaultWarningThreshold)
	result := Pass
	currentlyRunning := float32(app.TasksHealthy + app.TasksStaged)
	message := fmt.Sprintf("Only %d are healthy out of total %d", int(currentlyRunning), app.Instances)

	if currentlyRunning == 0.0 && app.Instances > 0 {
		result = Critical
	} else if currentlyRunning > 0.0 && currentlyRunning < failThreshold*float32(app.Instances) {
		result = Critical
	} else if currentlyRunning < warnThreshold*float32(app.Instances) {
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
