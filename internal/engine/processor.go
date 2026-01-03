package engine

import "log"

type Processor interface {
	ProcessEvent(relayID string, payload []byte) error
}

type Engine struct {
}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) ProcessEvent(relayID string, payload []byte) error {
	log.Printf("Processing Relay [%s] \n Payload: %s", relayID, string(payload))
	return nil
}
