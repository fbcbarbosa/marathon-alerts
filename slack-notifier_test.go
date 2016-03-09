package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResultToColor(t *testing.T) {
	slack := Slack{}
	assert.Equal(t, "good", *slack.resultToColor(Pass))
	assert.Equal(t, "good", *slack.resultToColor(Resolved))
	assert.Equal(t, "warning", *slack.resultToColor(Warning))
	assert.Equal(t, "danger", *slack.resultToColor(Critical))
	assert.Equal(t, "black", *slack.resultToColor(127))
}

func TestResultToString(t *testing.T) {
	slack := Slack{}
	assert.Equal(t, "Passed", slack.resultToString(Pass))
	assert.Equal(t, "Warning", slack.resultToString(Warning))
	assert.Equal(t, "Critical", slack.resultToString(Critical))
	assert.Equal(t, "Resolved", slack.resultToString(Resolved))
	assert.Equal(t, "Unknown", slack.resultToString(127))
}
