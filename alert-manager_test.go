package main

import (
	"testing"
	"time"

	"github.com/ashwanthkumar/marathon-alerts/checks"
	"github.com/ashwanthkumar/marathon-alerts/notifiers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestKey(t *testing.T) {
	check := checks.AppCheck{
		App:       "/foo",
		CheckName: "check-name",
	}
	mgr := AlertManager{}
	key := mgr.key(check, checks.Pass)
	assert.Equal(t, "/foo-check-name-99", key)
}

func TestKeyPrefix(t *testing.T) {
	check := checks.AppCheck{
		App:       "/foo",
		CheckName: "check-name",
	}
	mgr := AlertManager{}
	key := mgr.keyPrefix(check)
	assert.Equal(t, "/foo-check-name", key)
}

func TestCheckExists(t *testing.T) {
	suppressedApps := make(map[string]time.Time)
	suppressedApps["/foo-check-name-2"] = time.Now()
	check := checks.AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    checks.Warning,
	}
	mgr := AlertManager{
		AppSuppress: suppressedApps,
		AlertCount:  make(map[string]int),
	}
	exist, keyPrefixIfExist, keyIfExist, checkLevel := mgr.checkExist(check)
	assert.True(t, true, exist)
	assert.Equal(t, "/foo-check-name", keyPrefixIfExist)
	assert.Equal(t, "/foo-check-name-2", keyIfExist)
	assert.Equal(t, checks.Warning, checkLevel)
}

func TestProcessCheckWhenNewCheckArrives(t *testing.T) {
	suppressedApps := make(map[string]time.Time)
	mockNotifier := new(notifiers.MockNotifier)
	mockNotifier.On("Name").Return("mock-notifer")
	mockNotifier.On("Notify", mock.AnythingOfType("AppCheck")).Return(nil)
	check := checks.AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    checks.Warning,
	}
	mgr := AlertManager{
		AppSuppress: suppressedApps,
		AlertCount:  make(map[string]int),
		Notifiers:   []notifiers.Notifier{mockNotifier},
	}

	mgr.processCheck(check)
	expectedCheck := checks.AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    checks.Warning,
		Times:     1,
	}
	mockNotifier.AssertCalled(t, "Notify", expectedCheck)
	assert.Equal(t, mgr.AlertCount["/foo-check-name"], 1)
}

func TestProcessCheckWhenNewPassCheckArrives(t *testing.T) {
	mockNotifier := new(notifiers.MockNotifier)
	mockNotifier.On("Name").Return("mock-notifer")
	mockNotifier.On("Notify", mock.AnythingOfType("AppCheck")).Return(nil)
	suppressedApps := make(map[string]time.Time)
	check := checks.AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    checks.Pass,
	}
	alertCount := make(map[string]int)
	alertCount["/foo-check-name"] = 2
	mgr := AlertManager{
		AppSuppress: suppressedApps,
		AlertCount:  alertCount,
		Notifiers:   []notifiers.Notifier{mockNotifier},
	}

	mgr.processCheck(check)
	mockNotifier.AssertNotCalled(t, "Notify", check)
	assert.Equal(t, mgr.AlertCount["/foo-check-name"], 0)
}

func TestProcessCheckWhenExistingCheckOfDifferentLevel(t *testing.T) {
	mockNotifier := new(notifiers.MockNotifier)
	mockNotifier.On("Name").Return("mock-notifer")
	mockNotifier.On("Notify", mock.AnythingOfType("AppCheck")).Return(nil)

	suppressedApps := make(map[string]time.Time)
	suppressedApps["/foo-check-name-2"] = time.Now()
	check := checks.AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    checks.Critical,
	}
	alertCount := make(map[string]int)
	alertCount["/foo-check-name"] = 1
	mgr := AlertManager{
		AppSuppress: suppressedApps,
		AlertCount:  alertCount,
		Notifiers:   []notifiers.Notifier{mockNotifier},
	}

	mgr.processCheck(check)
	expectedCheck := checks.AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    checks.Critical,
		Times:     2,
	}
	mockNotifier.AssertCalled(t, "Notify", expectedCheck)
	assert.Equal(t, mgr.AlertCount["/foo-check-name"], 2)
}

func TestProcessCheckWhenExistingCheckOfSameLevel(t *testing.T) {
	mockNotifier := new(notifiers.MockNotifier)
	mockNotifier.On("Name").Return("mock-notifer")
	mockNotifier.On("Notify", mock.AnythingOfType("AppCheck")).Return(nil)

	suppressedApps := make(map[string]time.Time)
	suppressedApps["/foo-check-name-2"] = time.Now()
	alertCount := make(map[string]int)
	alertCount["/foo-check-name"] = 1
	check := checks.AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    checks.Warning,
	}
	mgr := AlertManager{
		AppSuppress: suppressedApps,
		AlertCount:  alertCount,
		Notifiers:   []notifiers.Notifier{mockNotifier},
	}

	mgr.processCheck(check)
	mockNotifier.AssertNotCalled(t, "Notify", check)
	assert.Equal(t, mgr.AlertCount["/foo-check-name"], 1)
}

func TestProcessCheckWhenResolvedCheckArrives(t *testing.T) {
	mockNotifier := new(notifiers.MockNotifier)
	mockNotifier.On("Name").Return("mock-notifer")
	mockNotifier.On("Notify", mock.AnythingOfType("AppCheck")).Return(nil)

	suppressedApps := make(map[string]time.Time)
	suppressedApps["/foo-check-name-2"] = time.Now()
	check := checks.AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    checks.Pass,
	}
	alertCount := make(map[string]int)
	alertCount["/foo-check-name"] = 1
	mgr := AlertManager{
		AppSuppress: suppressedApps,
		AlertCount:  alertCount,
		Notifiers:   []notifiers.Notifier{mockNotifier},
	}

	mgr.processCheck(check)
	expectedCheck := checks.AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    checks.Resolved,
		Times:     2,
	}
	mockNotifier.AssertCalled(t, "Notify", expectedCheck)
	// We remove AlertCount upon Resolved check
	assert.Equal(t, mgr.AlertCount["/foo-check-name"], 0)
}

func TestProcessCheckWhenNewCheckArrivesButDisabledViaLabels(t *testing.T) {
	mockNotifier := new(notifiers.MockNotifier)
	mockNotifier.On("Name").Return("mock-notifer")
	mockNotifier.On("Notify", mock.AnythingOfType("AppCheck")).Return(nil)

	suppressedApps := make(map[string]time.Time)
	appLabels := make(map[string]string)
	appLabels["alerts.enabled"] = "false"
	check := checks.AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    checks.Warning,
		Labels:    appLabels,
	}
	mgr := AlertManager{
		AppSuppress: suppressedApps,
		AlertCount:  make(map[string]int),
		Notifiers:   []notifiers.Notifier{mockNotifier},
	}

	mgr.processCheck(check)
	mockNotifier.AssertNotCalled(t, "Notify", check)
}

func TestCleanUpSupressedAlerts(t *testing.T) {
	suppressedApps := make(map[string]time.Time)
	suppressedApps["/foo-check-name-2"] = time.Now().Add(-5 * time.Minute)
	mgr := AlertManager{
		AppSuppress:      suppressedApps,
		AlertCount:       make(map[string]int),
		SuppressDuration: 1 * time.Minute,
	}

	assert.Equal(t, 1, len(mgr.AppSuppress))
	mgr.cleanUpSupressedAlerts()
	assert.Equal(t, 0, len(mgr.AppSuppress))
}

func TestCleanUpSupressedAlertsIgnoreIfLessThanSuppressDuration(t *testing.T) {
	suppressedApps := make(map[string]time.Time)
	suppressedApps["/foo-check-name-2"] = time.Now().Add(-5 * time.Minute)
	mgr := AlertManager{
		AppSuppress:      suppressedApps,
		AlertCount:       make(map[string]int),
		SuppressDuration: 10 * time.Minute,
	}

	assert.Equal(t, 1, len(mgr.AppSuppress))
	mgr.cleanUpSupressedAlerts()
	assert.Equal(t, 1, len(mgr.AppSuppress))
}

func TestTimesCountAfterTheCheckHasBeenIdleForSuppressedDuration(t *testing.T) {
	mockNotifier := new(notifiers.MockNotifier)
	mockNotifier.On("Name").Return("mock-notifer")
	mockNotifier.On("Notify", mock.AnythingOfType("AppCheck")).Return(nil)

	alertCount := make(map[string]int)
	alertCount["/foo-check-name"] = 1
	suppressedApps := make(map[string]time.Time)
	suppressedApps["/foo-check-name-2"] = time.Now().Add(-15 * time.Minute)
	mgr := AlertManager{
		AppSuppress:      suppressedApps,
		Notifiers:        []notifiers.Notifier{mockNotifier},
		AlertCount:       alertCount,
		SuppressDuration: 10 * time.Minute,
	}
	check := checks.AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    checks.Critical,
	}

	// When Times 1
	mgr.processCheck(check)
	assert.Equal(t, 2, mgr.AlertCount["/foo-check-name"])
	// After cleaning up supressed alerts
	mgr.cleanUpSupressedAlerts()
	mgr.processCheck(check)
	assert.Equal(t, 3, mgr.AlertCount["/foo-check-name"])
}
