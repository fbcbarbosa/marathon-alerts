package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestKey(t *testing.T) {
	check := AppCheck{
		App:       "/foo",
		CheckName: "check-name",
	}
	mgr := AlertManager{}
	key := mgr.key(check, Pass)
	assert.Equal(t, "/foo-check-name-99", key)
}

func TestCheckExists(t *testing.T) {
	suppressedApps := make(map[string]time.Time)
	suppressedApps["/foo-check-name-2"] = time.Now()
	check := AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    Warning,
	}
	mgr := AlertManager{
		AppSuppress: suppressedApps,
	}
	exist, keyIfExist, checkLevel := mgr.checkExist(check)
	assert.True(t, true, exist)
	assert.Equal(t, "/foo-check-name-2", keyIfExist)
	assert.Equal(t, Warning, checkLevel)
}

func TestProcessCheckWhenNewCheckArrives(t *testing.T) {
	notifierChannel := make(chan AppCheck, 1)
	suppressedApps := make(map[string]time.Time)
	check := AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    Warning,
	}
	mgr := AlertManager{
		AppSuppress:  suppressedApps,
		NotifierChan: notifierChannel,
	}

	mgr.processCheck(check)
	actualCheck := <-notifierChannel
	assert.Equal(t, "/foo", actualCheck.App)
	assert.Equal(t, "check-name", actualCheck.CheckName)
	assert.Equal(t, Warning, actualCheck.Result)
}

func TestProcessCheckWhenNewPassCheckArrives(t *testing.T) {
	notifierChannel := make(chan AppCheck, 1)
	suppressedApps := make(map[string]time.Time)
	check := AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    Pass,
	}
	mgr := AlertManager{
		AppSuppress:  suppressedApps,
		NotifierChan: notifierChannel,
	}

	mgr.processCheck(check)
	assert.Len(t, notifierChannel, 0)
}

func TestProcessCheckWhenExistingCheckOfDifferentLevel(t *testing.T) {
	notifierChannel := make(chan AppCheck, 1)
	suppressedApps := make(map[string]time.Time)
	suppressedApps["/foo-check-name-2"] = time.Now()
	check := AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    Critical,
	}
	mgr := AlertManager{
		AppSuppress:  suppressedApps,
		NotifierChan: notifierChannel,
	}

	mgr.processCheck(check)
	assert.Len(t, notifierChannel, 1)
	actualCheck := <-notifierChannel
	assert.Equal(t, "/foo", actualCheck.App)
	assert.Equal(t, "check-name", actualCheck.CheckName)
	assert.Equal(t, Critical, actualCheck.Result)
}

func TestProcessCheckWhenExistingCheckOfSameLevel(t *testing.T) {
	notifierChannel := make(chan AppCheck, 1)
	suppressedApps := make(map[string]time.Time)
	suppressedApps["/foo-check-name-2"] = time.Now()
	check := AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    Warning,
	}
	mgr := AlertManager{
		AppSuppress:  suppressedApps,
		NotifierChan: notifierChannel,
	}

	mgr.processCheck(check)
	assert.Len(t, notifierChannel, 0)
}

func TestProcessCheckWhenNewCheckArrivesButDisabledViaLabels(t *testing.T) {
	notifierChannel := make(chan AppCheck, 1)
	suppressedApps := make(map[string]time.Time)
	appLabels := make(map[string]string)
	appLabels["alerts.enabled"] = "false"
	check := AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    Warning,
		Labels:    appLabels,
	}
	mgr := AlertManager{
		AppSuppress:  suppressedApps,
		NotifierChan: notifierChannel,
	}

	mgr.processCheck(check)
	assert.Len(t, notifierChannel, 0)
}

func TestCleanUpSupressedAlerts(t *testing.T) {
	notifierChannel := make(chan AppCheck)
	suppressedApps := make(map[string]time.Time)
	suppressedApps["/foo-check-name-2"] = time.Now().Add(-5 * time.Minute)
	mgr := AlertManager{
		AppSuppress:      suppressedApps,
		NotifierChan:     notifierChannel,
		SuppressDuration: 1 * time.Minute,
	}

	assert.Equal(t, 1, len(mgr.AppSuppress))
	mgr.cleanUpSupressedAlerts()
	assert.Equal(t, 0, len(mgr.AppSuppress))
}

func TestCleanUpSupressedAlertsIgnoreIfLessThanSuppressDuration(t *testing.T) {
	notifierChannel := make(chan AppCheck)
	suppressedApps := make(map[string]time.Time)
	suppressedApps["/foo-check-name-2"] = time.Now().Add(-5 * time.Minute)
	mgr := AlertManager{
		AppSuppress:      suppressedApps,
		NotifierChan:     notifierChannel,
		SuppressDuration: 10 * time.Minute,
	}

	assert.Equal(t, 1, len(mgr.AppSuppress))
	mgr.cleanUpSupressedAlerts()
	assert.Equal(t, 1, len(mgr.AppSuppress))
}
