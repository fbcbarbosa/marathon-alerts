package main

import (
	"net/url"
	"testing"
	"time"

	"github.com/ashwanthkumar/marathon-alerts/checks"
	marathon "github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessCheckForAllSubscribers(t *testing.T) {
	appLabels := make(map[string]string)
	client := new(MockMarathon)
	apps := marathon.Applications{
		Apps: []marathon.Application{marathon.Application{Labels: appLabels}},
	}
	var urlValues url.Values
	client.On("Applications", urlValues).Return(&apps, nil)

	alertChan := make(chan checks.AppCheck, 1)
	check, now := CreateMockChecker(appLabels)

	appChecker := AppChecker{
		Client:        client,
		AlertsChannel: alertChan,
		Checks:        []checks.Checker{check},
	}

	expectedCheck := checks.AppCheck{Result: checks.Critical, Timestamp: now, App: "/foo-app", Labels: appLabels}
	err := appChecker.processChecks()
	assert.Nil(t, err)
	assert.Len(t, alertChan, 1)
	actualCheck := <-alertChan
	assert.Equal(t, actualCheck, expectedCheck)
}

func TestProcessCheckForWithNoSubscribers(t *testing.T) {
	appLabels := make(map[string]string)
	appLabels["alerts.checks.subscribe"] = "check-that-does-not-exist"
	client := new(MockMarathon)
	apps := marathon.Applications{
		Apps: []marathon.Application{marathon.Application{Labels: appLabels}},
	}
	var urlValues url.Values
	client.On("Applications", urlValues).Return(&apps, nil)

	alertChan := make(chan checks.AppCheck, 1)
	check, _ := CreateMockChecker(appLabels)

	appChecker := AppChecker{
		Client:        client,
		AlertsChannel: alertChan,
		Checks:        []checks.Checker{check},
	}

	err := appChecker.processChecks()
	assert.Nil(t, err)
	assert.Len(t, alertChan, 0)
}

func CreateMockChecker(appLabels map[string]string) (checks.Checker, time.Time) {
	now := time.Now()
	check := new(MockChecker)
	check.On("Name").Return("mock-check")
	check.On("Check", mock.AnythingOfType("Application")).Return(checks.AppCheck{
		Result:    checks.Critical,
		Timestamp: now,
		App:       "/foo-app",
		Labels:    appLabels,
	})

	return check, now
}
