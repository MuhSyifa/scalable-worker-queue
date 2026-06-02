package worker

import (
	"context"
	"fmt"
	"golang-worker-queue/internal/domain"
	"time"

	"github.com/rs/zerolog/log"
)

type JobHandler interface {
	Handle(ctx context.Context, job *domain.Job) error
}

type Processor struct {
	handlers map[domain.JobType]JobHandler
}

func NewProcessor() *Processor {
	return &Processor{
		handlers: make(map[domain.JobType]JobHandler),
	}
}

func (p *Processor) RegisterHandler(jobType domain.JobType, handler JobHandler) {
	p.handlers[jobType] = handler
}

func (p *Processor) Process(ctx context.Context, job *domain.Job) error {
	handler, exists := p.handlers[job.Type]
	if !exists {
		return fmt.Errorf("no handler registered for job type: %s", job.Type)
	}

	// Create a timeout context for the handler
	handleCtx, cancel := context.WithTimeout(ctx, 30*time.Minute) // Default max duration
	defer cancel()

	log.Debug().Msgf("Executing handler for job %s of type %s", job.ID, job.Type)
	return handler.Handle(handleCtx, job)
}
