package engine

import (
	"context"
	"log"
	"sync"
	"time"

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
	ctx        context.Context
	cancel     context.CancelFunc
}

// Constructor with dependency injxtn
func NewWorkerPool(maxWorkers int, db *store.Store, reg *Registry) *WorkerPool {
	return &WorkerPool{
		JobQueue:   make(chan Job, 100),
		MaxWorkers: maxWorkers,
		Store:      db,
		Registry:   reg,
	}
}

// Spaws all worker goroutines
func (wp *WorkerPool) Start(ctx context.Context) {
	wp.ctx, wp.cancel = context.WithCancel(ctx)
	log.Printf("Starting worker pool with %d workers", wp.MaxWorkers)
	for i := 0; i < wp.MaxWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
	log.Printf("All %d workers started and ready", wp.MaxWorkers)
}

// Each worker runs its own goroutine
func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()
	log.Printf("Worker %d started", id)
	for {
		select {
		case <-wp.ctx.Done():
			log.Printf("Worker %d shutting down", id)
			return
		case job := <-wp.JobQueue:
			start := time.Now()
			log.Printf("Worker %d processing relay %s", id, job.RelayID)
			err := wp.process(wp.ctx, job)
			duration := time.Since(start)
			if err != nil {
				log.Printf("Worker %d failed relay %s in %v: %v", id, job.RelayID, duration, err)
				job.MsgAck(false)
			} else {
				log.Printf("Worker %d finished relay %s in %v", id, job.RelayID, duration)
				job.MsgAck(true)
			}
		}
	}
}

// Executes the actual workflow logic
func (wp *WorkerPool) process(ctx context.Context, job Job) (err error) {
	status := "success"
	details := "Relay executed successfully"
	defer func() {
		if err != nil {
			status = "failed"
			details = err.Error()
		}
		logErr := wp.Store.LogExecution(context.Background(), job.RelayID, status, details)
		if logErr != nil {
			log.Printf("Failed to save execution log: %v", logErr)
		}
	}()

	instruction, fetchErr := wp.Store.GetRelayInstructions(ctx, job.RelayID)
	if fetchErr != nil {
		return fetchErr
	}

	executor, pluginErr := wp.Registry.Get(instruction.ActionType)
	if pluginErr != nil {
		return pluginErr
	}

	return executor.Execute(ctx, instruction.Config, job.Payload)
}

func (wp *WorkerPool) Shutdown() {
	log.Printf("Initializing worker pool shutdown")

	if wp.cancel != nil {
		wp.cancel()
	}
	close(wp.JobQueue)
	wp.wg.Wait()
	log.Printf("Worker pool shutdown complete")
}
