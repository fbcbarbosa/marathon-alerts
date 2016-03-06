package main

import (
	"fmt"
	"sync"
)

type Notifier interface {
	Notify(check AppCheck)
	Name() string
}

type NotifyManager struct {
	// channel to get checks for notification
	AlertChan chan AppCheck
	Notifiers []Notifier

	RunWaitGroup      sync.WaitGroup
	NotifierWaitGroup sync.WaitGroup
	stopChannel       chan bool
}

func (n *NotifyManager) Start() {
	fmt.Println("Starting Notify Manager...")
	n.RunWaitGroup.Add(1)
	n.stopChannel = make(chan bool)
	go n.run()
	fmt.Println("Notify Manager Started.")
}

func (n *NotifyManager) Stop() {
	fmt.Println("Stopping Notify Manager...")
	close(n.stopChannel)
	n.RunWaitGroup.Done()
}

func (n *NotifyManager) Notify(check AppCheck) {
	n.NotifierWaitGroup.Add(1)
	// Send the notifications for check
	for _, notifier := range n.Notifiers {
		notifier.Notify(check)
	}
	n.NotifierWaitGroup.Done()
}

func (n *NotifyManager) run() {
	running := true
	for running {
		select {
		case check := <-n.AlertChan:
			n.NotifierWaitGroup.Wait()
			n.Notify(check)
		case <-n.stopChannel:
			running = false
		}
	}
}
