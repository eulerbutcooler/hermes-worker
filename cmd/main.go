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
	reg := engine.NewRegistry()
	reg.Register("debug_log", debug.New())
	reg.Register("discord_send", discord.New())
	log.Println("Integrations loaded: debug_log, discord_send")
	ctx, cancel := context.WithCancel(context.Background())
	pool := engine.NewWorkerPool(10, db, reg)
	pool.Start(ctx)
	eng := engine.NewEngine()
	consumer, err := queue.NewConsumer(cfg.NatsURL, eng)
	if err != nil {
		log.Fatalf("NATS Init Error: %v", err)
	}
	consumer.Start()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down worker...")
	cancel()
	pool.Shutdown()
}
