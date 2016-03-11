package main

import (
	"fmt"
	"sync"
	"time"

	maps "github.com/ashwanthkumar/golang-utils/maps"
	"github.com/ashwanthkumar/marathon-alerts/checks"
	"github.com/ashwanthkumar/marathon-alerts/notifiers"
	"github.com/ashwanthkumar/marathon-alerts/routes"
	"github.com/rcrowley/go-metrics"
)

const (
	AlertsEnabledLabel = "alerts.enabled"
	AppRoutesLabel     = "alerts.routes"
)

type AlertManager struct {
	CheckerChan      chan checks.AppCheck // channel to get app check results
	AppSuppress      map[string]time.Time // Key - AppName-CheckName-CheckResult
	AlertCount       map[string]int       // Key - AppName-CheckName -> Consecutive # of failures
	SuppressDuration time.Duration
	Notifiers        []notifiers.Notifier
	RunWaitGroup     sync.WaitGroup
	stopChannel      chan bool
	supressMutex     sync.Mutex
}

func (a *AlertManager) Start() {
	fmt.Println("Starting Alert Manager...")
	a.RunWaitGroup.Add(1)
	a.stopChannel = make(chan bool)
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

func (a *AlertManager) processCheck(check checks.AppCheck) {
	a.supressMutex.Lock()
	defer a.supressMutex.Unlock()

	alertEnabled := maps.GetBoolean(check.Labels, AlertsEnabledLabel, true)

	if alertEnabled {
		allRoutes, err := routes.ParseRoutes(maps.GetString(check.Labels, AppRoutesLabel, routes.DefaultRoutes))
		if err != nil {
			fmt.Printf("Error - %v\n", err)
			return
		}
		checkExists, keyPrefixIfCheckExists, keyIfCheckExists, resultIfCheckExists := a.checkExist(check)

		if checkExists && check.Result == checks.Pass {
			a.AlertCount[keyPrefixIfCheckExists]++
			check.Times = a.AlertCount[keyPrefixIfCheckExists]
			check.Result = checks.Resolved
			delete(a.AppSuppress, keyIfCheckExists)
			delete(a.AlertCount, keyPrefixIfCheckExists)
			a.notifyCheck(check, allRoutes)
			a.incNotifCounter(check)
		} else if checkExists && check.Result != resultIfCheckExists {
			delete(a.AppSuppress, keyIfCheckExists)
			key := a.key(check, check.Result)
			a.AppSuppress[key] = check.Timestamp
			a.AlertCount[keyPrefixIfCheckExists]++
			check.Times = a.AlertCount[keyPrefixIfCheckExists]
			a.notifyCheck(check, allRoutes)
			a.incNotifCounter(check)
		} else if !checkExists && check.Result != checks.Pass {
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
			a.notifyCheck(check, allRoutes)
			a.incNotifCounter(check)
		} else if !checkExists && check.Result == checks.Pass {
			keyPrefix := a.keyPrefix(check)
			delete(a.AlertCount, keyPrefix)
		}
	} else {
		fmt.Printf("Monitoring disabled for %s via alerts.enabled label in app config\n", check.App)
	}
}

func (a *AlertManager) notifyCheck(check checks.AppCheck, allRoutes []routes.Route) {
	for _, route := range allRoutes {
		if route.Match(check) {
			for _, notifier := range a.Notifiers {
				if route.MatchNotifier(notifier.Name()) {
					notifier.Notify(check)
				}
			}
		}
	}
}

func (a *AlertManager) checkExist(check checks.AppCheck) (bool, string, string, checks.CheckStatus) {
	for _, level := range checks.CheckLevels {
		keyPrefix := a.keyPrefix(check)
		key := a.key(check, level)
		_, present := a.AppSuppress[key]
		if present {
			return true, keyPrefix, key, level
		}
	}

	return false, "", "", checks.Pass
}

func (a *AlertManager) key(check checks.AppCheck, level checks.CheckStatus) string {
	return fmt.Sprintf("%s-%d", a.keyPrefix(check), level)
}

func (a *AlertManager) keyPrefix(check checks.AppCheck) string {
	return fmt.Sprintf("%s-%s", check.App, check.CheckName)
}

func (a *AlertManager) incNotifCounter(check checks.AppCheck) {
	metrics.GetOrRegisterCounter("notifications-total", nil).Inc(1)
	metrics.GetOrRegisterMeter("notifications-rate", nil).Mark(1)
	if check.Result == checks.Warning {
		metrics.GetOrRegisterCounter("notifications-warning", nil).Inc(1)
		metrics.GetOrRegisterMeter("notifications-warning-rate", DebugMetricsRegistry).Mark(1)
	} else if check.Result == checks.Critical {
		metrics.GetOrRegisterCounter("notifications-critical", nil).Inc(1)
		metrics.GetOrRegisterMeter("notifications-critical-rate", DebugMetricsRegistry).Mark(1)
	} else if check.Result == checks.Pass {
		metrics.GetOrRegisterCounter("notifications-pass", nil).Inc(1)
		metrics.GetOrRegisterMeter("notifications-pass-rate", DebugMetricsRegistry).Mark(1)
	} else if check.Result == checks.Resolved {
		metrics.GetOrRegisterCounter("notifications-resolved", nil).Inc(1)
		metrics.GetOrRegisterMeter("notifications-resolved-rate", DebugMetricsRegistry).Mark(1)
	} else {
		panic("Calling incCheckCounter for " + fmt.Sprintf("%v", check))
	}
}
