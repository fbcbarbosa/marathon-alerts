package main

import (
	"sync"
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
	notifierChannel := make(chan AppCheck)
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

	appCheckAssertion := func(t *testing.T, check AppCheck) {
		assert.Equal(t, "/foo", check.App)
		assert.Equal(t, "check-name", check.CheckName)
		assert.Equal(t, Warning, check.Result)
	}

	testWG := AssertOnChannel(t, notifierChannel, 1*time.Second, appCheckAssertion)
	mgr.processCheck(check)
	testWG.Wait()
}

func TestProcessCheckWhenExistingCheckOfDifferentLevel(t *testing.T) {
	notifierChannel := make(chan AppCheck)
	suppressedApps := make(map[string]time.Time)
	suppressedApps["/foo-check-name-2"] = time.Now()
	check := AppCheck{
		App:       "/foo",
		CheckName: "check-name",
		Result:    Fail,
	}
	mgr := AlertManager{
		AppSuppress:  suppressedApps,
		NotifierChan: notifierChannel,
	}

	assertCalled := false
	appCheckAssertion := func(t *testing.T, check AppCheck) {
		assert.Equal(t, "/foo", check.App)
		assert.Equal(t, "check-name", check.CheckName)
		assert.Equal(t, Fail, check.Result)
		assertCalled = true
	}

	testWG := AssertOnChannel(t, notifierChannel, 1*time.Second, appCheckAssertion)
	mgr.processCheck(check)
	testWG.Wait()
	assert.True(t, assertCalled)
}

func TestProcessCheckWhenExistingCheckOfSameLevel(t *testing.T) {
	notifierChannel := make(chan AppCheck)
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

	assertCalled := false
	appCheckAssertion := func(t *testing.T, check AppCheck) {
		assertCalled = true
	}

	testWG := AssertOnChannel(t, notifierChannel, 1*time.Second, appCheckAssertion)
	mgr.processCheck(check)
	testWG.Wait()

	assert.False(t, assertCalled)
}

func AssertOnChannel(t *testing.T, channel chan AppCheck, timeout time.Duration, assert func(*testing.T, AppCheck)) sync.WaitGroup {
	var wg sync.WaitGroup
	go func(t *testing.T, channel chan AppCheck, wg sync.WaitGroup, timeout time.Duration, assert func(*testing.T, AppCheck)) {
		running := true
		wg.Add(1)
		for running {
			select {
			case checkToAssert := <-channel:
				assert(t, checkToAssert)
				running = false
				wg.Done()
			case <-time.After(timeout):
				running = false
				wg.Done()
			}
		}
	}(t, channel, wg, timeout, assert)

	return wg
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
