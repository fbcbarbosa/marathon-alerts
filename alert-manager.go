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

	checkExists, keyIfCheckExists, levelIfCheckExists := a.checkExist(check)

	if checkExists && check.Result == Pass {
		a.NotifierChan <- check
		delete(a.AppSuppress, keyIfCheckExists)
	} else if checkExists && check.Result != levelIfCheckExists {
		a.NotifierChan <- check
		delete(a.AppSuppress, keyIfCheckExists)
		key := a.key(check, check.Result)
		a.AppSuppress[key] = time.Now()
	} else if !checkExists && check.Result != Pass {
		a.NotifierChan <- check
		key := a.key(check, check.Result)
		a.AppSuppress[key] = time.Now()
	} else {
		// same check of same level - ignore until cleanUpSupressedAlerts cleans the existing check
	}

	a.supressMutex.Unlock()
}

func (a *AlertManager) checkExist(check AppCheck) (bool, string, CheckStatus) {
	for _, level := range CheckLevels {
		key := a.key(check, level)
		_, present := a.AppSuppress[key]
		if present {
			return true, key, level
		}
	}

	return false, "", Pass
}

func (a *AlertManager) key(check AppCheck, level CheckStatus) string {
	return fmt.Sprintf("%s-%s-%d", check.App, check.CheckName, level)
}
