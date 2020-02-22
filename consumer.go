package main

import (
	cloudevents "github.com/cloudevents/sdk-go"
)

type Consumer struct {
	labels []string
	client cloudevents.Client
}

func (c Consumer) getLabels() []string {
	return c.labels
}

func (c *Consumer) Notify() {
	// TODO: create event and send
}
