package main

import (
	"fmt"
	"sync"
	"time"

	maps "github.com/ashwanthkumar/golang-utils/maps"
	"github.com/rcrowley/go-metrics"
)

const AlertsEnabledLabel = "alerts.enabled"

type AlertManager struct {
	CheckerChan      chan AppCheck        // channel to get app check results
	NotifierChan     chan AppCheck        // channel to send app check notifications
	AppSuppress      map[string]time.Time // Key - AppName-CheckName-CheckResult
	AlertCount       map[string]int       // Key - AppName-CheckName -> Consecutive # of failures
	SuppressDuration time.Duration
	RunWaitGroup     sync.WaitGroup
	stopChannel      chan bool
	supressMutex     sync.Mutex
}

func (a *AlertManager) Start() {
	fmt.Println("Starting Alert Manager...")
	a.RunWaitGroup.Add(1)
	a.stopChannel = make(chan bool)
	a.NotifierChan = make(chan AppCheck)
	a.AppSuppress = make(map[string]time.Time)
	a.AlertCount = make(map[string]int)
	go a.run()
	fmt.Println("Alert Manager Started.")
}

func (a *AlertManager) Stop() {
	fmt.Println("Stopping Alert Manager...")
	close(a.stopChannel)
	a.RunWaitGroup.Done()
}

func (a *AlertManager) cleanUpSupressedAlerts() {
	a.supressMutex.Lock()
	for key, suppressedOn := range a.AppSuppress {
		if time.Now().Sub(suppressedOn) > a.SuppressDuration {
			metrics.GetOrRegisterCounter("alerts-suppressed-cleaned", nil).Inc(int64(1))
			delete(a.AppSuppress, key)
		}
	}
	a.supressMutex.Unlock()
}

func (a *AlertManager) run() {
	running := true
	for running {
		select {
		case <-time.After(5 * time.Second):
			metrics.GetOrRegisterCounter("alerts-suppressed-called", DebugMetricsRegistry).Inc(int64(1))
			a.cleanUpSupressedAlerts()
		case check := <-a.CheckerChan:
			metrics.GetOrRegisterCounter("alerts-process-check-called", DebugMetricsRegistry).Inc(int64(1))
			a.processCheck(check)
		case <-a.stopChannel:
			metrics.GetOrRegisterCounter("alerts-manager-stopped", DebugMetricsRegistry).Inc(int64(1))
			running = false
		}
	}
}

func (a *AlertManager) processCheck(check AppCheck) {
	a.supressMutex.Lock()
	defer a.supressMutex.Unlock()

	alertEnabled := maps.GetBoolean(check.Labels, AlertsEnabledLabel, true)

	if alertEnabled {
		checkExists, keyPrefixIfCheckExists, keyIfCheckExists, resultIfCheckExists := a.checkExist(check)

		if checkExists && check.Result == Pass {
			delete(a.AppSuppress, keyIfCheckExists)
			delete(a.AlertCount, keyPrefixIfCheckExists)
			a.NotifierChan <- check
			a.incNotifCounter(check)
		} else if checkExists && check.Result != resultIfCheckExists {
			delete(a.AppSuppress, keyIfCheckExists)
			key := a.key(check, check.Result)
			a.AppSuppress[key] = check.Timestamp
			a.AlertCount[keyPrefixIfCheckExists]++
			check.Times = a.AlertCount[keyPrefixIfCheckExists]
			a.NotifierChan <- check
			a.incNotifCounter(check)
		} else if !checkExists && check.Result != Pass {
			keyPrefix := a.keyPrefix(check)
			key := a.key(check, check.Result)
			a.AppSuppress[key] = check.Timestamp
			_, present := a.AlertCount[keyPrefix]
			if present {
				a.AlertCount[keyPrefix]++
			} else {
				a.AlertCount[keyPrefix] = 1
			}
			check.Times = a.AlertCount[keyPrefix]
			a.NotifierChan <- check
			a.incNotifCounter(check)
		} else if !checkExists && check.Result == Pass {
			keyPrefix := a.keyPrefix(check)
			delete(a.AlertCount, keyPrefix)
		}
	} else {
		fmt.Printf("Monitoring disabled for %s via alerts.enabled label in app config\n", check.App)
	}
}

func (a *AlertManager) checkExist(check AppCheck) (bool, string, string, CheckStatus) {
	for _, level := range CheckLevels {
		keyPrefix := a.keyPrefix(check)
		key := a.key(check, level)
		_, present := a.AppSuppress[key]
		if present {
			return true, keyPrefix, key, level
		}
	}

	return false, "", "", Pass
}

func (a *AlertManager) key(check AppCheck, level CheckStatus) string {
	return fmt.Sprintf("%s-%d", a.keyPrefix(check), level)
}

func (a *AlertManager) keyPrefix(check AppCheck) string {
	return fmt.Sprintf("%s-%s", check.App, check.CheckName)
}

func (a *AlertManager) incNotifCounter(check AppCheck) {
	metrics.GetOrRegisterCounter("notifications-total", nil).Inc(1)
	metrics.GetOrRegisterMeter("notifications-rate", nil).Mark(1)
	if check.Result == Warning {
		metrics.GetOrRegisterCounter("notifications-warning", nil).Inc(1)
		metrics.GetOrRegisterMeter("notifications-warning-rate", DebugMetricsRegistry).Mark(1)
	} else if check.Result == Critical {
		metrics.GetOrRegisterCounter("notifications-critical", nil).Inc(1)
		metrics.GetOrRegisterMeter("notifications-critical-rate", DebugMetricsRegistry).Mark(1)
	} else if check.Result == Pass {
		metrics.GetOrRegisterCounter("notifications-resolved", nil).Inc(1)
		metrics.GetOrRegisterMeter("notifications-resolved-rate", DebugMetricsRegistry).Mark(1)
	} else {
		panic("Calling incCheckCounter for " + fmt.Sprintf("%v", check))
	}
}
