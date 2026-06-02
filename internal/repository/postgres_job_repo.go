package repository

import (
	"golang-worker-queue/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PostgresJobRepository struct {
	db *sqlx.DB
}

func NewPostgresJobRepository(db *sqlx.DB) domain.JobRepository {
	return &PostgresJobRepository{db: db}
}

func (r *PostgresJobRepository) CreateJob(job *domain.Job) error {
	query := `
		INSERT INTO jobs (id, type, payload, status, priority, max_retries, retries, scheduled_at, created_at, updated_at)
		VALUES (:id, :type, :payload, :status, :priority, :max_retries, :retries, :scheduled_at, :created_at, :updated_at)
	`
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	_, err := r.db.NamedExec(query, job)
	return err
}

func (r *PostgresJobRepository) GetJob(id uuid.UUID) (*domain.Job, error) {
	var job domain.Job
	query := `SELECT * FROM jobs WHERE id = $1`
	err := r.db.Get(&job, query, id)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *PostgresJobRepository) UpdateJobStatus(id uuid.UUID, status domain.JobStatus, lastError error) error {
	var errStr *string
	if lastError != nil {
		s := lastError.Error()
		errStr = &s
	}

	query := `
		UPDATE jobs
		SET status = $1, last_error = $2, updated_at = $3
		WHERE id = $4
	`
	_, err := r.db.Exec(query, status, errStr, time.Now(), id)
	return err
}

func (r *PostgresJobRepository) ListJobs(status *domain.JobStatus, limit, offset int) ([]domain.Job, error) {
	var jobs []domain.Job
	var err error
	if status != nil {
		query := `SELECT * FROM jobs WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		err = r.db.Select(&jobs, query, *status, limit, offset)
	} else {
		query := `SELECT * FROM jobs ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		err = r.db.Select(&jobs, query, limit, offset)
	}
	return jobs, err
}

func (r *PostgresJobRepository) CancelJob(id uuid.UUID) error {
	query := `
		UPDATE jobs
		SET status = $1, updated_at = $2
		WHERE id = $3 AND status IN ($4, $5)
	`
	_, err := r.db.Exec(query, domain.StatusCancelled, time.Now(), id, domain.StatusPending, domain.StatusRetrying)
	return err
}
