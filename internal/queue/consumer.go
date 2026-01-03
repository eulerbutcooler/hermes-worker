package queue

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/eulerbutcooler/hermes-worker/internal/engine"
	"github.com/nats-io/nats.go"
)

type Consumer struct {
	js        nats.JetStream
	processor engine.Processor
}

func NewConsumer(url string, processor engine.Processor) (*Consumer, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("nats connect error: %w", err)
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetsream init error: %w", err)
	}
	return &Consumer{
		js:        js,
		processor: processor,
	}, nil
}

func (c *Consumer) Start() {
	_, err := c.js.Subscribe("events.>", func(msg *nats.Msg) {
		msg.Ack()
		type Event struct {
			RelayID string `json:"relay_id"`
		}
		var evt Event
		if err := json.Unmarshal(msg.Data, &evt); err != nil {
			log.Printf("Error parsing JSON: %v", err)
			return
		}
		err := c.processor.ProcessEvent(evt.RelayID, msg.Data)
		if err != nil {
			log.Printf("Failed to process event: %v", err)
		}
	}, nats.Durable("WORKER_CONSUMER"), nats.ManualAck())
	if err != nil {
		log.Fatalf("Subscription failed: %v", err)
	}
	log.Printf("Worker is listening for events...")
}
