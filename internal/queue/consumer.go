package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/eulerbutcooler/hermes-worker/internal/engine"
	"github.com/nats-io/nats.go"
)

type Consumer struct {
	js       nats.JetStream
	sub      *nats.Subscription
	jobQueue chan engine.Job
}

// Constructor pattern
// Initializes the NATS connection but doesnt start consuming right off
func NewConsumer(url string, jobQueue chan engine.Job) (*Consumer, error) {
	nc, err := nats.Connect(url, nats.MaxReconnects(10), nats.ReconnectWait(2))
	if err != nil {
		return nil, fmt.Errorf("nats connect error: %w", err)
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetsream init error: %w", err)
	}
	return &Consumer{
		js:       js,
		jobQueue: jobQueue,
	}, nil
}

// Consumes the messages by subscibing to NATS and processing messages async
func (c *Consumer) Start() error {
	sub, err := c.js.Subscribe("events.>", c.handleMessage, nats.Durable("WORKER_CONSUMER"), nats.ManualAck(), nats.AckWait(30*time.Second))
	if err != nil {
		return fmt.Errorf("subscription failed: %w", err)
	}
	c.sub = sub
	log.Printf("Worker consumer started, listening for events...")
	return nil
}

func (c *Consumer) handleMessage(msg *nats.Msg) {
	msg.Ack()
	type Event struct {
		RelayID    string          `json:"relay_id"`
		Payload    json.RawMessage `json:"payload"`
		ReceivedAt string          `json:"received_at"`
	}
	var evt Event
	if err := json.Unmarshal(msg.Data, &evt); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		msg.Nak()
		return
	}
	log.Printf("Received event for relay: %s", evt.RelayID)

	// Bridges NATS consumer to Worker Pool
	job := engine.Job{
		RelayID: evt.RelayID,
		Payload: evt.Payload,
		MsgAck: func(success bool) {
			if success {
				msg.Ack()
				log.Printf("Acknowledged message for relay: %s", evt.RelayID)
			} else {
				msg.Nak()
				log.Printf("Nacked message for relay %s (requeued the message for retry)", evt.RelayID)
			}
		},
	}
	//Blocking send to channel - If the worker is full this will wait
	c.jobQueue <- job
}

func (c *Consumer) Stop() error {
	if c.sub != nil {
		// To process remaining messages and then close
		return c.sub.Drain()
	}
	return nil
}
