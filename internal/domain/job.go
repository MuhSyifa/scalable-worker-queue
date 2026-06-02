package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusRunning   JobStatus = "running"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
	StatusRetrying  JobStatus = "retrying"
	StatusCancelled JobStatus = "cancelled"
	StatusDLQ       JobStatus = "dlq"
)

type JobType string

const (
	JobTypeEmail       JobType = "email"
	JobTypeWebhook     JobType = "webhook"
	JobTypePayment     JobType = "payment_callback"
	JobTypeReport      JobType = "report_generation"
	JobTypeNotification JobType = "notification"
	JobTypeImageProcess JobType = "image_processing"
)

type Job struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	Type        JobType         `json:"type" db:"type"`
	Payload     json.RawMessage `json:"payload" db:"payload"`
	Status      JobStatus       `json:"status" db:"status"`
	Priority    int             `json:"priority" db:"priority"`
	MaxRetries  int             `json:"max_retries" db:"max_retries"`
	Retries     int             `json:"retries" db:"retries"`
	LastError   *string         `json:"last_error" db:"last_error"`
	ScheduledAt *time.Time      `json:"scheduled_at" db:"scheduled_at"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

type JobLog struct {
	ID        uuid.UUID `json:"id" db:"id"`
	JobID     uuid.UUID `json:"job_id" db:"job_id"`
	Status    JobStatus `json:"status" db:"status"`
	Message   string    `json:"message" db:"message"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type WorkerLog struct {
	ID        uuid.UUID `json:"id" db:"id"`
	WorkerID  string    `json:"worker_id" db:"worker_id"`
	Status    string    `json:"status" db:"status"`
	Message   string    `json:"message" db:"message"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// JobRepository interface for abstracting database/queue operations
type JobRepository interface {
	CreateJob(job *Job) error
	GetJob(id uuid.UUID) (*Job, error)
	UpdateJobStatus(id uuid.UUID, status JobStatus, lastError error) error
	ListJobs(status *JobStatus, limit, offset int) ([]Job, error)
	CancelJob(id uuid.UUID) error
}
