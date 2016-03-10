package checks

import (
	"fmt"
	"time"

	"github.com/gambol99/go-marathon"
)

type SuspendedCheck struct{}

func (s *SuspendedCheck) Name() string {
	return "suspended"
}

func (n *SuspendedCheck) Check(app marathon.Application) AppCheck {
	var result CheckStatus
	var message string
	if app.Instances == 0 {
		result = Critical
		message = fmt.Sprintf("%s is suspended.", app.ID)
	} else {
		result = Pass
		message = fmt.Sprintf("%s is not suspended.", app.ID)
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
