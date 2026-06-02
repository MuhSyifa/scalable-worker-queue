package handlers

import (
	"context"
	"encoding/json"
	"golang-worker-queue/internal/domain"
	"time"

	"github.com/rs/zerolog/log"
)

type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type EmailHandler struct{}

func NewEmailHandler() *EmailHandler {
	return &EmailHandler{}
}

func (h *EmailHandler) Handle(ctx context.Context, job *domain.Job) error {
	var payload EmailPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return err
	}

	log.Info().Msgf("Sending email to %s with subject: %s", payload.To, payload.Subject)
	
	// Simulate work
	time.Sleep(2 * time.Second)

	log.Info().Msgf("Email sent successfully to %s", payload.To)
	return nil
}
