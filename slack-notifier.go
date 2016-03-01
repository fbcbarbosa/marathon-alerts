package main

import (
	"fmt"
	"strings"

	"github.com/ashwanthkumar/slack-go-webhook"
)

type Slack struct {
	Webhook string
	Channel string
	Owners  []string
}

func (s *Slack) Notify(check AppCheck) {
	attachment := slack.Attachment{
		Text:  &check.Message,
		Color: s.resultToColor(check),
	}
	attachment.
		AddField(slack.Field{Title: "App", Value: check.App, Short: true}).
		AddField(slack.Field{Title: "Check", Value: check.CheckName, Short: true}).
		AddField(slack.Field{Title: "Result", Value: s.resultToString(check), Short: true})

	var destination string
	if s.Channel != "" {
		destination = s.Channel
	}

	mainText := ""
	if s.Owners != nil && len(s.Owners) > 0 {
		mainText = mainText + "Hey " + s.parseOwners(s.Owners) + ", Please check!"
	}

	payload := slack.Payload(mainText,
		"marathon-alerts",
		"",
		destination,
		[]slack.Attachment{attachment})

	err := slack.Send(s.Webhook, payload)
	if err != nil {
		fmt.Printf("Unexpected Error - %v", err)
	}
}

func (s *Slack) resultToColor(check AppCheck) *string {
	color := "black"
	switch check.Result {
	case Pass:
		color = "good"
	case Warning:
		color = "warning"
	case Fail:
		color = "danger"
	}

	return &color
}

func (s *Slack) resultToString(check AppCheck) string {
	value := "Unknown"
	switch check.Result {
	case Pass:
		value = "Passed"
	case Warning:
		value = "Warning"
	case Fail:
		value = "Failed"
	}

	return value
}

func (s *Slack) parseOwners(owners []string) string {
	parsedOwners := []string{}
	for _, owner := range owners {
		parsedOwners = append(parsedOwners, fmt.Sprintf("@%s", owner))
	}

	return strings.Join(parsedOwners, ", ")
}
