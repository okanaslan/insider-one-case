package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"insider-one-case/internal/idempotency"
	"insider-one-case/internal/model"
	"insider-one-case/internal/repository"
	appvalidator "insider-one-case/internal/validator"
)

var ErrDuplicateEvent = errors.New("event already processed")

type EventService struct {
	repo          *repository.EventRepository
	idempotency   *idempotency.RedisStore
	eventValidate *appvalidator.EventValidator
	log           *slog.Logger
}

func NewEventService(
	repo *repository.EventRepository,
	idempotency *idempotency.RedisStore,
	eventValidate *appvalidator.EventValidator,
	log *slog.Logger,
) *EventService {
	return &EventService{
		repo:          repo,
		idempotency:   idempotency,
		eventValidate: eventValidate,
		log:           log,
	}
}

func (s *EventService) Ingest(ctx context.Context, req model.EventIngestRequest) (model.EventIngestResponse, error) {
	if err := s.eventValidate.ValidateEvent(ctx, req); err != nil {
		return model.EventIngestResponse{}, err
	}

	if req.EventID == "" {
		req.EventID = uuid.NewString()
	}

	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now().UTC()
	}

	reserved, err := s.idempotency.ReserveEvent(ctx, "event:"+req.EventID, 24*time.Hour)
	if err != nil {
		s.log.Warn("idempotency reserve failed; proceeding", "event_id", req.EventID, "error", err)
	} else if !reserved {
		return model.EventIngestResponse{}, ErrDuplicateEvent
	}

	// TODO: enqueue to async worker pipeline and persist batch to ClickHouse.
	if err := s.repo.InsertEvent(ctx, req); err != nil {
		return model.EventIngestResponse{}, err
	}

	return model.EventIngestResponse{
		EventID: req.EventID,
		Status:  "accepted",
	}, nil
}
