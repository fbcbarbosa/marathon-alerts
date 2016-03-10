package main

import (
	"testing"
	"time"

	"github.com/ashwanthkumar/marathon-alerts/checks"
	"github.com/stretchr/testify/assert"
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
	notifierChannel := make(chan checks.AppCheck, 1)
	suppressedApps := make(map[string]time.Time)
	check := checks.AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    checks.Warning,
	}
	mgr := AlertManager{
		AppSuppress:  suppressedApps,
		AlertCount:   make(map[string]int),
		NotifierChan: notifierChannel,
	}

	mgr.processCheck(check)
	actualCheck := <-notifierChannel
	assert.Equal(t, "/foo", actualCheck.App)
	assert.Equal(t, "check-name", actualCheck.CheckName)
	assert.Equal(t, checks.Warning, actualCheck.Result)
	assert.Equal(t, 1, actualCheck.Times)
	assert.Equal(t, mgr.AlertCount["/foo-check-name"], 1)
}

func TestProcessCheckWhenNewPassCheckArrives(t *testing.T) {
	notifierChannel := make(chan checks.AppCheck, 1)
	suppressedApps := make(map[string]time.Time)
	check := checks.AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    checks.Pass,
	}
	alertCount := make(map[string]int)
	alertCount["/foo-check-name"] = 2
	mgr := AlertManager{
		AppSuppress:  suppressedApps,
		AlertCount:   alertCount,
		NotifierChan: notifierChannel,
	}

	mgr.processCheck(check)
	assert.Len(t, notifierChannel, 0)
	assert.Equal(t, mgr.AlertCount["/foo-check-name"], 0)
}

func TestProcessCheckWhenExistingCheckOfDifferentLevel(t *testing.T) {
	notifierChannel := make(chan checks.AppCheck, 1)
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
		AppSuppress:  suppressedApps,
		AlertCount:   alertCount,
		NotifierChan: notifierChannel,
	}

	mgr.processCheck(check)
	assert.Len(t, notifierChannel, 1)
	actualCheck := <-notifierChannel
	assert.Equal(t, "/foo", actualCheck.App)
	assert.Equal(t, "check-name", actualCheck.CheckName)
	assert.Equal(t, checks.Critical, actualCheck.Result)
	assert.Equal(t, 2, actualCheck.Times)
	assert.Equal(t, mgr.AlertCount["/foo-check-name"], 2)
}

func TestProcessCheckWhenExistingCheckOfSameLevel(t *testing.T) {
	notifierChannel := make(chan checks.AppCheck, 1)
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
		AppSuppress:  suppressedApps,
		AlertCount:   alertCount,
		NotifierChan: notifierChannel,
	}

	mgr.processCheck(check)
	assert.Len(t, notifierChannel, 0)
	assert.Equal(t, mgr.AlertCount["/foo-check-name"], 1)
}

func TestProcessCheckWhenResolvedCheckArrives(t *testing.T) {
	notifierChannel := make(chan checks.AppCheck, 1)
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
		AppSuppress:  suppressedApps,
		AlertCount:   alertCount,
		NotifierChan: notifierChannel,
	}

	mgr.processCheck(check)
	assert.Len(t, notifierChannel, 1)
	actualCheck := <-notifierChannel
	assert.Equal(t, "/foo", actualCheck.App)
	assert.Equal(t, "check-name", actualCheck.CheckName)
	assert.Equal(t, checks.Resolved, actualCheck.Result)
	assert.Equal(t, 2, actualCheck.Times)
	// We remove AlertCount upon Resolved check
	assert.Equal(t, mgr.AlertCount["/foo-check-name"], 0)
}

func TestProcessCheckWhenNewCheckArrivesButDisabledViaLabels(t *testing.T) {
	notifierChannel := make(chan checks.AppCheck, 1)
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
		AppSuppress:  suppressedApps,
		AlertCount:   make(map[string]int),
		NotifierChan: notifierChannel,
	}

	mgr.processCheck(check)
	assert.Len(t, notifierChannel, 0)
}

func TestCleanUpSupressedAlerts(t *testing.T) {
	notifierChannel := make(chan checks.AppCheck)
	suppressedApps := make(map[string]time.Time)
	suppressedApps["/foo-check-name-2"] = time.Now().Add(-5 * time.Minute)
	mgr := AlertManager{
		AppSuppress:      suppressedApps,
		NotifierChan:     notifierChannel,
		AlertCount:       make(map[string]int),
		SuppressDuration: 1 * time.Minute,
	}

	assert.Equal(t, 1, len(mgr.AppSuppress))
	mgr.cleanUpSupressedAlerts()
	assert.Equal(t, 0, len(mgr.AppSuppress))
}

func TestCleanUpSupressedAlertsIgnoreIfLessThanSuppressDuration(t *testing.T) {
	notifierChannel := make(chan checks.AppCheck)
	suppressedApps := make(map[string]time.Time)
	suppressedApps["/foo-check-name-2"] = time.Now().Add(-5 * time.Minute)
	mgr := AlertManager{
		AppSuppress:      suppressedApps,
		NotifierChan:     notifierChannel,
		AlertCount:       make(map[string]int),
		SuppressDuration: 10 * time.Minute,
	}

	assert.Equal(t, 1, len(mgr.AppSuppress))
	mgr.cleanUpSupressedAlerts()
	assert.Equal(t, 1, len(mgr.AppSuppress))
}

func TestTimesCountAfterTheCheckHasBeenIdleForSuppressedDuration(t *testing.T) {
	notifierChannel := make(chan checks.AppCheck, 2)
	alertCount := make(map[string]int)
	alertCount["/foo-check-name"] = 1
	suppressedApps := make(map[string]time.Time)
	suppressedApps["/foo-check-name-2"] = time.Now().Add(-15 * time.Minute)
	mgr := AlertManager{
		AppSuppress:      suppressedApps,
		NotifierChan:     notifierChannel,
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
