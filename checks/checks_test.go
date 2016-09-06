package checks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResultToString(t *testing.T) {
	assert.Equal(t, "Passed", CheckStatusToString(Pass))
	assert.Equal(t, "Warning", CheckStatusToString(Warning))
	assert.Equal(t, "Critical", CheckStatusToString(Critical))
	assert.Equal(t, "Resolved", CheckStatusToString(Resolved))
	assert.Equal(t, "Unknown", CheckStatusToString(127))
}
