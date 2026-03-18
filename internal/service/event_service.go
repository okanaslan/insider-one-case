package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"insider-one-case/internal/idempotency"
	"insider-one-case/internal/model"
)

var (
	ErrDuplicateEvent = errors.New("event already processed")
	ErrEnqueueFailed  = errors.New("failed to enqueue event")
)

// EventEnqueuer is the interface the service uses to submit events for async processing.
type EventEnqueuer interface {
	Enqueue(model.EventIngestRequest) error
}

type EventService struct {
	queue       EventEnqueuer
	idempotency *idempotency.RedisStore
	log         *slog.Logger
}

func NewEventService(
	queue EventEnqueuer,
	idempotency *idempotency.RedisStore,
	log *slog.Logger,
) *EventService {
	return &EventService{
		queue:       queue,
		idempotency: idempotency,
		log:         log,
	}
}

func (s *EventService) Ingest(ctx context.Context, req model.EventIngestRequest) (model.EventIngestResponse, error) {
	key := "event:" + req.UniquenessKey()

	reserved, err := s.idempotency.ReserveEvent(ctx, key, 24*time.Hour)
	if err != nil {
		s.log.Warn("idempotency reserve failed; proceeding", "uniqueness_key", req.UniquenessKey(), "error", err)
	} else if !reserved {
		return model.EventIngestResponse{}, ErrDuplicateEvent
	}

	if err := s.queue.Enqueue(req); err != nil {
		s.log.Error("failed to enqueue event", "uniqueness_key", req.UniquenessKey(), "error", err)
		return model.EventIngestResponse{}, ErrEnqueueFailed
	}

	return model.EventIngestResponse{
		Status:  "accepted",
		Message: "event accepted for processing",
	}, nil
}
