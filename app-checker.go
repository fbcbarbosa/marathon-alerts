package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	marathon "github.com/gambol99/go-marathon"
)

type CheckStatus uint8

const (
	Pass    = CheckStatus(99)
	Warning = CheckStatus(2)
	Fail    = CheckStatus(1)
)

var CheckLevels = [...]CheckStatus{Warning, Fail}

type AppChecker struct {
	Client        marathon.Marathon // client to access the marathon instance
	RunWaitGroup  sync.WaitGroup
	CheckInterval time.Duration // --poll-interval
	stopChannel   chan bool
	Checks        []Checker
	// NB: move thhis to notifier // Supress       time.Duration // --supress-duration
	AlertsChannel chan AppCheck
}

func (a *AppChecker) Start() {
	fmt.Println("Starting App Checker...")
	a.RunWaitGroup.Add(1)
	a.stopChannel = make(chan bool)
	a.AlertsChannel = make(chan AppCheck)
	go a.run()
}

func (a *AppChecker) Stop() {
	fmt.Println("Stopping App Checker...")
	close(a.stopChannel)
	a.RunWaitGroup.Done()
}

func (a *AppChecker) run() {
	running := true
	for running {
		select {
		case <-time.After(a.CheckInterval):
			err := a.processChecks()
			if err != nil {
				log.Fatalf("Unexpected error - %v\n", err)
			}
		case <-a.stopChannel:
			running = false
		}
		time.Sleep(1 * time.Second)
	}
}

func (a *AppChecker) processChecks() error {
	println("checking all the apps")
	apps, err := a.Client.Applications(nil)
	if err != nil {
		return err
	}
	for _, app := range apps.Apps {
		for _, check := range a.Checks {
			result := check.Check(app)
			a.AlertsChannel <- result
		}
	}

	return nil
}
