package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/eulerbutcooler/hermes-worker/internal/config"
	"github.com/eulerbutcooler/hermes-worker/internal/engine"
	"github.com/eulerbutcooler/hermes-worker/internal/integrations/debug"
	"github.com/eulerbutcooler/hermes-worker/internal/integrations/discord"
	"github.com/eulerbutcooler/hermes-worker/internal/queue"
	"github.com/eulerbutcooler/hermes-worker/internal/store"
)

func main() {
	cfg := config.LoadConfig()
	db, err := store.NewStore(cfg.DbURL)
	if err != nil {
		log.Fatalf("DB Init Error: %v", err)
	}
	log.Printf("DB Connected")

	//Registry Pattern
	// Registering integrations instead of hardcoding
	reg := engine.NewRegistry()
	reg.Register("debug_log", debug.New())
	reg.Register("discord_send", discord.New())
	log.Println("Integrations loaded: debug_log, discord_send")

	pool := engine.NewWorkerPool(10, db, reg)
	ctx, cancel := context.WithCancel(context.Background())
	pool.Start(ctx)

	consumer, err := queue.NewConsumer(cfg.NatsURL, pool.JobQueue)
	if err != nil {
		log.Fatalf("NATS Init Error: %v", err)
	}
	if err := consumer.Start(); err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	if err := consumer.Stop(); err != nil {
		log.Printf("Error stopping consumer: %v", err)
	}
	log.Println("Shutting down worker...")
	cancel()
	pool.Shutdown()
	log.Println("Worker stoppped gracefully")
}
