package service

import (
	"context"
	"golang-worker-queue/internal/domain"
	"golang-worker-queue/internal/queue"

	"github.com/google/uuid"
)

type JobService struct {
	repo       domain.JobRepository
	redisQueue *queue.RedisQueue
}

func NewJobService(repo domain.JobRepository, redisQueue *queue.RedisQueue) *JobService {
	return &JobService{
		repo:       repo,
		redisQueue: redisQueue,
	}
}

func (s *JobService) CreateJob(ctx context.Context, job *domain.Job) error {
	job.ID = uuid.New()
	job.Status = domain.StatusPending
	
	if job.MaxRetries == 0 {
		job.MaxRetries = 3
	}

	err := s.repo.CreateJob(job)
	if err != nil {
		return err
	}

	// Enqueue job to Redis
	return s.redisQueue.Enqueue(ctx, job)
}

func (s *JobService) GetJob(id uuid.UUID) (*domain.Job, error) {
	return s.repo.GetJob(id)
}

func (s *JobService) ListJobs(status *domain.JobStatus, limit, offset int) ([]domain.Job, error) {
	return s.repo.ListJobs(status, limit, offset)
}

func (s *JobService) CancelJob(id uuid.UUID) error {
	return s.repo.CancelJob(id)
}
