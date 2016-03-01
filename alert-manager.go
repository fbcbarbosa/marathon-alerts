package main

import (
	"fmt"
	"sync"
	"time"
)

type AlertManager struct {
	// channel to get app check results
	CheckerChan chan AppCheck
	// channel to send app check notifications
	NotifierChan chan AppCheck
	// Key - AppName-CheckName-CheckResult
	AppSuppress      map[string]time.Time
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
	go a.run()
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
			a.cleanUpSupressedAlerts()
		case check := <-a.CheckerChan:
			a.processCheck(check)
		case <-a.stopChannel:
			running = false
		}
	}
}

func (a *AlertManager) processCheck(check AppCheck) {
	a.supressMutex.Lock()
	// TODO - Fix this - WE need to send a notification on Pass if we earlier had failed
	if check.Result != Pass {
		fmt.Printf("%s has failed %s check\n", check.App, check.CheckName)
		keyPrefix := fmt.Sprintf("%s-%s", check.App, check.CheckName)
		suppressed, lastKnownLevel := a.isCheckSuppressed(check, keyPrefix)

		if !suppressed {
			a.NotifierChan <- check
			key := fmt.Sprintf("%s-%s", keyPrefix, lastKnownLevel)
			delete(a.AppSuppress, key)
			key = fmt.Sprintf("%s-%s", keyPrefix, check.Result)
			a.AppSuppress[key] = time.Now()
		}
	}
	a.supressMutex.Unlock()
}

func (a *AlertManager) isCheckSuppressed(check AppCheck, keyPrefix string) (bool, CheckStatus) {
	for _, level := range CheckLevels {
		key := fmt.Sprintf("%s-%s", keyPrefix, level)
		lastNoticiedTime, present := a.AppSuppress[key]
		if present && check.Result != level {
			return false, level
		} else if present && check.Result == level {
			suppressed := time.Now().Sub(lastNoticiedTime) < a.SuppressDuration
			return suppressed, level
		}
	}

	return false, check.Result
}
