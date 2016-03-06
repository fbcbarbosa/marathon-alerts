package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	maps "github.com/ashwanthkumar/golang-utils/maps"
	sets "github.com/ashwanthkumar/golang-utils/sets"
	marathon "github.com/gambol99/go-marathon"
)

const (
	CheckSubscriptionLabel = "alerts.checks.subscribe"
	SubscribeAllChecks     = "all"
)

type CheckStatus uint8

const (
	Pass     = CheckStatus(99)
	Warning  = CheckStatus(2)
	Critical = CheckStatus(1)
)

var CheckLevels = [...]CheckStatus{Warning, Critical}

type AppChecker struct {
	Client        marathon.Marathon
	RunWaitGroup  sync.WaitGroup
	CheckInterval time.Duration
	stopChannel   chan bool
	Checks        []Checker
	AlertsChannel chan AppCheck
	// Snooze the entire system for some Time
	// Useful if we don't want to SPAM the notifications
	// when doing maintenance of mesos cluster
	// TODO - Enable this feature via API endpoint
	IsSnoozed  bool
	SnoozedAt  time.Time
	SnoozedFor time.Duration
}

func (a *AppChecker) Start() {
	fmt.Println("Starting App Checker...")
	a.RunWaitGroup.Add(1)
	a.stopChannel = make(chan bool)
	a.AlertsChannel = make(chan AppCheck)

	a.IsSnoozed = false

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
		checksSubscribed := sets.FromSlice(
			strings.Split(maps.GetString(app.Labels, CheckSubscriptionLabel, SubscribeAllChecks),
				","))
		for _, check := range a.Checks {
			if checksSubscribed.Contains(check.Name()) || checksSubscribed.Contains(SubscribeAllChecks) {
				result := check.Check(app)
				a.AlertsChannel <- result
			}
		}
	}

	return nil
}
