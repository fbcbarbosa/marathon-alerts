package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	marathon "github.com/gambol99/go-marathon"
)

var appChecker AppChecker
var alertManager AlertManager
var notifyManager NotifyManager

// Check settings
var minHealthyWarningThreshold float32
var minHealthyFailThreshold float32

// Required flags
var marathonURI string
var checkInterval time.Duration
var alertSuppressDuration time.Duration

// Slack flags
var slackWebhooks []string
var slackChannel string
var slackOwners string

func main() {
	os.Args[0] = "marathon-alerts"
	defineFlags()
	flag.Parse()
	client, err := marathonClient(marathonURI)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	minHealthyTasks := &MinHealthyTasks{
		DefaultFailThreshold:    minHealthyFailThreshold,
		DefaultWarningThreshold: minHealthyWarningThreshold,
	}
	checks := []Checker{minHealthyTasks}

	appChecker = AppChecker{
		Client:        client,
		CheckInterval: checkInterval,
		Checks:        checks,
	}
	appChecker.Start()

	alertManager = AlertManager{
		CheckerChan:      appChecker.AlertsChannel,
		SuppressDuration: alertSuppressDuration,
	}
	alertManager.Start()

	slackOwners := strings.Split(slackOwners, ",")
	if len(slackOwners) < 1 {
		slackOwners = []string{}
	}

	var notifiers []Notifier
	for _, slackTeam := range slackWebhooks {
		slack := Slack{
			Webhook: slackTeam,
			Channel: slackChannel,
			Owners:  slackOwners,
		}
		notifiers = append(notifiers, &slack)
	}
	notifyManager = NotifyManager{
		AlertChan: alertManager.NotifierChan,
		Notifiers: notifiers,
	}
	notifyManager.Start()

	appChecker.RunWaitGroup.Wait()
	// Handle signals and cleanup all routines
}

func marathonClient(uri string) (marathon.Marathon, error) {
	config := marathon.NewDefaultConfig()
	config.URL = uri
	config.HTTPClient = &http.Client{
		Timeout: (30 * time.Second),
	}

	return marathon.NewClient(config)
}

func defineFlags() {
	flag.StringVar(&marathonURI, "uri", "", "Marathon URI to connect")
	flag.DurationVar(&checkInterval, "check-interval", 30*time.Second, "Check runs periodically on this interval")
	flag.DurationVar(&alertSuppressDuration, "alerts-suppress-duration", 30*time.Minute, "Suppress alerts for this duration once notified")

	// Check flags
	flag.Float32Var(&minHealthyWarningThreshold, "check-min-healthy-warn-threshold", 0.8, "Min instances check warning threshold")
	flag.Float32Var(&minHealthyFailThreshold, "check-min-healthy-fail-threshold", 0.6, "Min instances check fail threshold")

	// Slack flags
	flag.StringSliceVar(&slackWebhooks, "slack-webhooks", []string{}, "Slack webhook to post the alert")
	flag.StringVar(&slackChannel, "slack-channel", "", "#Channel / @User to post the alert (defaults to webhook configuration)")
	flag.StringVar(&slackOwners, "slack-owner", "", "Comma list of owners who should be alerted on the post")
}
