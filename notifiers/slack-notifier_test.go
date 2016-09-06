package notifiers

import (
	"testing"

	"github.com/ashwanthkumar/marathon-alerts/checks"
	"github.com/stretchr/testify/assert"
)

func TestResultToColor(t *testing.T) {
	slack := Slack{}
	assert.Equal(t, "good", *slack.resultToColor(checks.Pass))
	assert.Equal(t, "good", *slack.resultToColor(checks.Resolved))
	assert.Equal(t, "warning", *slack.resultToColor(checks.Warning))
	assert.Equal(t, "danger", *slack.resultToColor(checks.Critical))
	assert.Equal(t, "black", *slack.resultToColor(127))
}
