package config

import (
	"log"
	"os"
)

type Config struct {
	NatsURL string
	DbURL   string
}

func LoadConfig() *Config {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@loclahost:5432/hermes"
	}

	log.Printf("Loaded Config: NatsURL=%s", natsURL)
	return &Config{
		NatsURL: natsURL,
		DbURL:   dbURL,
	}
}
