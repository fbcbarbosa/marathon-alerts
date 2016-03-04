package main

import (
	"log"
	"net/url"
	"testing"
	"time"

	"github.com/ashwanthkumar/marathon-alerts/mocks"
	marathon "github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessCheckForAllSubscribers(t *testing.T) {
	appLabels := make(map[string]string)
	client := new(mocks.Marathon)
	apps := marathon.Applications{
		Apps: []marathon.Application{marathon.Application{Labels: appLabels}},
	}
	var urlValues url.Values
	client.On("Applications", urlValues).Return(&apps, nil)

	alertChan := make(chan AppCheck)
	check, now := CreateMockChecker(appLabels)

	appChecker := AppChecker{
		Client:        client,
		AlertsChannel: alertChan,
		Checks:        []Checker{check},
	}

	expectedCheck := AppCheck{Result: Critical, Timestamp: now, App: "/foo-app", Labels: appLabels}
	wg := AssertOnChannel(t, alertChan, 5*time.Second, func(t *testing.T, actualData AppCheck) {
		assert.Equal(t, expectedCheck, actualData)
	})
	err := appChecker.processChecks()
	assert.Nil(t, err)
	wg.Wait()
}

func TestProcessCheckForWithNoSubscribers(t *testing.T) {
	appLabels := make(map[string]string)
	appLabels["alerts.checks.subscribe"] = "check-that-does-not-exist"
	client := new(mocks.Marathon)
	apps := marathon.Applications{
		Apps: []marathon.Application{marathon.Application{Labels: appLabels}},
	}
	var urlValues url.Values
	client.On("Applications", urlValues).Return(&apps, nil)

	alertChan := make(chan AppCheck)
	check, _ := CreateMockChecker(appLabels)

	appChecker := AppChecker{
		Client:        client,
		AlertsChannel: alertChan,
		Checks:        []Checker{check},
	}

	wg := AssertOnChannel(t, alertChan, 5*time.Second, func(t *testing.T, actualData AppCheck) {
		log.Println("Should not be called")
		t.FailNow()
	})
	err := appChecker.processChecks()
	assert.Nil(t, err)
	wg.Wait()
}

func CreateMockChecker(appLabels map[string]string) (Checker, time.Time) {
	now := time.Now()
	check := new(MockChecker)
	check.On("Name").Return("mock-check")
	check.On("Check", mock.AnythingOfType("Application")).Return(AppCheck{
		Result:    Critical,
		Timestamp: now,
		App:       "/foo-app",
		Labels:    appLabels,
	})

	return check, now
}
