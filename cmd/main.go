package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/eulerbutcooler/hermes-worker/internal/config"
	"github.com/eulerbutcooler/hermes-worker/internal/engine"
	"github.com/eulerbutcooler/hermes-worker/internal/queue"
)

func main() {
	cfg := config.LoadConfig()
	eng := engine.NewEngine()
	consumer, err := queue.NewConsumer(cfg.NatsURL, eng)
	if err != nil {
		log.Fatalf("Failed to start NATS consumer: %v", err)
	}
	consumer.Start()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down worker...")
}
