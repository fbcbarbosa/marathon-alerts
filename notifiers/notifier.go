package notifiers

import "github.com/ashwanthkumar/marathon-alerts/checks"

type Notifier interface {
	Notify(check checks.AppCheck)
	Name() string
}
