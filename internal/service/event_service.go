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
	ErrOverloaded     = errors.New("service overloaded")
)

// EventEnqueuer is the interface the service uses to submit events for async processing.
type EventEnqueuer interface {
	Enqueue(ctx context.Context, event model.EventIngestRequest) error
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

	if err := s.queue.Enqueue(ctx, req); err != nil {
		s.log.Error("failed to enqueue event", "uniqueness_key", req.UniquenessKey(), "error", err)
		return model.EventIngestResponse{}, ErrOverloaded
	}

	return model.EventIngestResponse{
		Status:  "accepted",
		Message: "event accepted for processing",
	}, nil
}

// IngestBulk processes multiple events with per-event validation and partial-success semantics.
func (s *EventService) IngestBulk(ctx context.Context, req model.BulkEventIngestRequest) model.BulkEventIngestResponse {
	summary := model.BulkEventIngestSummary{
		Total: len(req.Events),
	}

	for _, event := range req.Events {
		key := "event:" + event.UniquenessKey()

		// Try to reserve the event key in Redis.
		reserved, err := s.idempotency.ReserveEvent(ctx, key, 24*time.Hour)
		if err != nil {
			s.log.Warn("idempotency reserve failed in bulk", "uniqueness_key", event.UniquenessKey(), "error", err)
		} else if !reserved {
			summary.Duplicate++
			continue
		}

		// Attempt to enqueue the event.
		if err := s.queue.Enqueue(ctx, event); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				summary.Overloaded++
			} else {
				s.log.Error("failed to enqueue event in bulk", "uniqueness_key", event.UniquenessKey(), "error", err)
				summary.Error++
			}
			continue
		}

		summary.Accepted++
	}

	// Determine overall response status.
	respStatus := "accepted_all"
	if summary.Accepted < summary.Total {
		respStatus = "accepted_partial"
	}

	return model.BulkEventIngestResponse{
		Status:  respStatus,
		Summary: summary,
	}
}
