package worker

import (
	"context"
	"fmt"
	"golang-worker-queue/internal/domain"
	"golang-worker-queue/internal/queue"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type WorkerPool struct {
	concurrency int
	queues      []string
	redisQueue  *queue.RedisQueue
	jobRepo     domain.JobRepository
	processor   *Processor
	wg          sync.WaitGroup
}

func NewWorkerPool(concurrency int, redisQueue *queue.RedisQueue, jobRepo domain.JobRepository, processor *Processor) *WorkerPool {
	// Define queue priorities (0 is highest)
	queues := []string{
		fmt.Sprintf("%s0", queue.QueuePrefix),
		fmt.Sprintf("%s1", queue.QueuePrefix),
		fmt.Sprintf("%s2", queue.QueuePrefix),
	}

	return &WorkerPool{
		concurrency: concurrency,
		queues:      queues,
		redisQueue:  redisQueue,
		jobRepo:     jobRepo,
		processor:   processor,
	}
}

func (p *WorkerPool) Start(ctx context.Context) {
	log.Info().Msgf("Starting worker pool with concurrency %d", p.concurrency)

	for i := 0; i < p.concurrency; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}
}

func (p *WorkerPool) Stop() {
	log.Info().Msg("Stopping worker pool, waiting for active jobs to finish...")
	p.wg.Wait()
	log.Info().Msg("Worker pool stopped gracefully.")
}

func (p *WorkerPool) worker(ctx context.Context, workerID int) {
	defer p.wg.Done()
	log.Debug().Msgf("Worker %d started", workerID)

	for {
		select {
		case <-ctx.Done():
			log.Debug().Msgf("Worker %d stopping due to context cancellation", workerID)
			return
		default:
			job, err := p.redisQueue.Dequeue(ctx, p.queues, 5*time.Second)
			if err != nil {
				log.Error().Err(err).Msgf("Worker %d failed to dequeue job", workerID)
				time.Sleep(1 * time.Second)
				continue
			}

			if job == nil {
				// Timeout, continue loop
				continue
			}

			log.Info().Msgf("Worker %d processing job %s of type %s", workerID, job.ID, job.Type)
			p.processJob(ctx, job)
		}
	}
}

func (p *WorkerPool) processJob(ctx context.Context, job *domain.Job) {
	// Mark job as running in DB
	err := p.jobRepo.UpdateJobStatus(job.ID, domain.StatusRunning, nil)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to update job %s to running", job.ID)
	}

	// Execute job
	err = p.processor.Process(ctx, job)

	if err != nil {
		log.Error().Err(err).Msgf("Job %s failed", job.ID)
		
		// Handle Retry or DLQ
		job.Retries++
		if job.Retries >= job.MaxRetries {
			log.Warn().Msgf("Job %s max retries reached, moving to DLQ", job.ID)
			p.jobRepo.UpdateJobStatus(job.ID, domain.StatusDLQ, err)
			// Ideally push to a Redis DLQ here as well
		} else {
			log.Info().Msgf("Retrying job %s (attempt %d/%d)", job.ID, job.Retries, job.MaxRetries)
			p.jobRepo.UpdateJobStatus(job.ID, domain.StatusRetrying, err)
			
			// Exponential backoff
			backoff := time.Duration(job.Retries*job.Retries) * time.Second
			job.ScheduledAt = func() *time.Time { t := time.Now().Add(backoff); return &t }()
			p.redisQueue.Enqueue(ctx, job)
		}
	} else {
		log.Info().Msgf("Job %s completed successfully", job.ID)
		p.jobRepo.UpdateJobStatus(job.ID, domain.StatusCompleted, nil)
	}
}
