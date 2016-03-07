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
	"github.com/rcrowley/go-metrics"
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
			metrics.GetOrRegisterCounter("apps-checker-stopped", DebugMetricsRegistry).Inc(int64(1))
			running = false
		}
		time.Sleep(1 * time.Second)
	}
}

func (a *AppChecker) processChecks() error {
	var apps *marathon.Applications
	var err error
	metrics.GetOrRegisterTimer("marathon-all-apps-response-time", nil).Time(func() {
		apps, err = a.Client.Applications(nil)
	})
	metrics.GetOrRegisterCounter("apps-checker-marathon-all-apps-api", DebugMetricsRegistry).Inc(int64(1))
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
				metrics.GetOrRegisterCounter("apps-checker-alerts-sent", DebugMetricsRegistry).Inc(int64(1))
				metrics.GetOrRegisterCounter("apps-checker-check-"+check.Name(), DebugMetricsRegistry).Inc(int64(1))
				metrics.GetOrRegisterCounter("apps-checker-app-"+app.ID, DebugMetricsRegistry).Inc(int64(1))
				metrics.GetOrRegisterCounter("apps-checker-"+app.ID+"-"+check.Name(), DebugMetricsRegistry).Inc(int64(1))
			}
		}
	}

	return nil
}
