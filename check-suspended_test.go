package main

import (
	"testing"

	"github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
)

func TestSuspendedCheckWhenEverythingIsFine(t *testing.T) {
	check := SuspendedCheck{}
	app := marathon.Application{
		ID:        "/foo",
		Instances: 10,
	}

	appCheck := check.Check(app)
	assert.Equal(t, Pass, appCheck.Result)
	assert.Equal(t, "suspended", appCheck.CheckName)
	assert.Equal(t, "/foo", appCheck.App)
	assert.Equal(t, "/foo is not suspended.", appCheck.Message)
}

func TestSuspendedCheckWhenAppIsSuspended(t *testing.T) {
	check := SuspendedCheck{}
	app := marathon.Application{
		ID:        "/foo",
		Instances: 0,
	}

	appCheck := check.Check(app)
	assert.Equal(t, Critical, appCheck.Result)
	assert.Equal(t, "suspended", appCheck.CheckName)
	assert.Equal(t, "/foo", appCheck.App)
	assert.Equal(t, "/foo is suspended.", appCheck.Message)
}
