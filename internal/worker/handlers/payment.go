package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"golang-worker-queue/internal/domain"
	"math/rand"
	"time"

	"github.com/rs/zerolog/log"
)

type PaymentPayload struct {
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"`
}

type PaymentHandler struct{}

func NewPaymentHandler() *PaymentHandler {
	return &PaymentHandler{}
}

func (h *PaymentHandler) Handle(ctx context.Context, job *domain.Job) error {
	var payload PaymentPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return err
	}

	log.Info().Msgf("Processing payment callback for transaction %s", payload.TransactionID)
	
	time.Sleep(1 * time.Second)

	// Simulate random failure to demonstrate retries
	if rand.Intn(100) < 30 {
		return fmt.Errorf("third party payment API timeout for transaction %s", payload.TransactionID)
	}

	log.Info().Msgf("Successfully processed payment callback for %s", payload.TransactionID)
	return nil
}
