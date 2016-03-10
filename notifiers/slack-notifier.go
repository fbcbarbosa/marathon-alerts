package notifiers

import (
	"fmt"
	"strings"

	maps "github.com/ashwanthkumar/golang-utils/maps"
	"github.com/ashwanthkumar/marathon-alerts/checks"
	"github.com/ashwanthkumar/slack-go-webhook"
)

type Slack struct {
	Webhook string
	Channel string
	Owners  string
}

func (s *Slack) Name() string {
	return "slack"
}

func (s *Slack) Notify(check checks.AppCheck) {
	attachment := slack.Attachment{
		Text:  &check.Message,
		Color: s.resultToColor(check.Result),
	}
	attachment.
		AddField(slack.Field{Title: "App", Value: check.App, Short: true}).
		AddField(slack.Field{Title: "Check", Value: check.CheckName, Short: true}).
		AddField(slack.Field{Title: "Result", Value: s.resultToString(check.Result), Short: true}).
		AddField(slack.Field{Title: "Times", Value: fmt.Sprintf("%d", check.Times), Short: true})

	destination := maps.GetString(check.Labels, "alerts.slack.channel", s.Channel)

	appSpecificOwners := maps.GetString(check.Labels, "alerts.slack.owners", s.Owners)
	var owners []string
	if appSpecificOwners != "" {
		owners = strings.Split(appSpecificOwners, ",")
	} else {
		owners = []string{"@here"}
	}

	alertSuffix := "Please check!"
	if check.Result == checks.Resolved {
		alertSuffix = "Check Resolved, thanks!"
	} else if check.Result == checks.Pass {
		alertSuffix = "Check Passed"
	}
	mainText := fmt.Sprintf("%s, %s", s.parseOwners(owners), alertSuffix)

	payload := slack.Payload(mainText,
		"marathon-alerts",
		"",
		destination,
		[]slack.Attachment{attachment})

	webhooks := strings.Split(maps.GetString(check.Labels, "alerts.slack.webhook", s.Webhook), ",")

	for _, webhook := range webhooks {
		err := slack.Send(webhook, "", payload)
		if err != nil {
			fmt.Printf("Unexpected Error - %v", err)
		}
	}
}

func (s *Slack) resultToColor(result checks.CheckStatus) *string {
	color := "black"
	switch {
	case checks.Pass == result || checks.Resolved == result:
		color = "good"
	case checks.Warning == result:
		color = "warning"
	case checks.Critical == result:
		color = "danger"
	}

	return &color
}

func (s *Slack) resultToString(result checks.CheckStatus) string {
	value := "Unknown"
	switch result {
	case checks.Pass:
		value = "Passed"
	case checks.Resolved:
		value = "Resolved"
	case checks.Warning:
		value = "Warning"
	case checks.Critical:
		value = "Critical"
	}

	return value
}

func (s *Slack) parseOwners(owners []string) string {
	parsedOwners := []string{}
	for _, owner := range owners {
		if owner != "@here" {
			owner = fmt.Sprintf("@%s", owner)
		}
		parsedOwners = append(parsedOwners, owner)
	}

	return strings.Join(parsedOwners, ", ")
}
