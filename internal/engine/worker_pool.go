package engine

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/eulerbutcooler/hermes-worker/internal/store"
)

type Job struct {
	RelayID string
	Payload []byte
	MsgAck  func(bool)
}

type WorkerPool struct {
	JobQueue   chan Job
	MaxWorkers int
	Store      *store.Store
	Registry   *Registry
	wg         sync.WaitGroup
}

func NewWorkerPool(maxWorkers int, db *store.Store, reg *Registry) *WorkerPool {
	return &WorkerPool{
		JobQueue:   make(chan Job, 100),
		MaxWorkers: maxWorkers,
		Store:      db,
		Registry:   reg,
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.MaxWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
}

func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()
	log.Printf("Worker %d started", id)
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-wp.JobQueue:
			err := wp.process(ctx, job)
			if err != nil {
				log.Printf("Worker %d failed relay %s: %v", id, job.RelayID, err)
				job.MsgAck(false)
			} else {
				log.Printf("Worker %d finished relay %s", id, job.RelayID)
				job.MsgAck(true)
			}
		}
	}
}

func (wp *WorkerPool) process(ctx context.Context, job Job) error {
	instruction, err := wp.Store.GetRelayInstructions(ctx, job.RelayID)
	if err != nil {
		if errors.Is(err, store.ErrRelayNotFound) {
			log.Printf("Dropping event for unknown/inactive relay: %s", job.RelayID)
			return nil
		}
		return err
	}
	executor, err := wp.Registry.Get(instruction.ActionType)
	if err != nil {
		log.Printf("Unknown action type '%s' for relay %s. Dropping.", instruction.ActionType, job.RelayID)
		return nil
	}
	return executor.Execute(ctx, instruction.Config, job.Payload)
}

func (wp *WorkerPool) Shutdown() {
	close(wp.JobQueue)
	wp.wg.Wait()
}
