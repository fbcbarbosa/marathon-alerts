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
	Client        marathon.Marathon
	RunWaitGroup  sync.WaitGroup
	CheckInterval time.Duration
	stopChannel   chan bool
	Checks        []Checker
	AlertsChannel chan AppCheck
}

func (a *AppChecker) Start() {
	fmt.Println("Starting App Checker...")
	a.RunWaitGroup.Add(1)
	a.stopChannel = make(chan bool)
	a.AlertsChannel = make(chan AppCheck)
	go a.run()
	fmt.Println("App Checker Started.")
	fmt.Printf("App Checker - Checking the status of all the apps every %v\n", a.CheckInterval)
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
