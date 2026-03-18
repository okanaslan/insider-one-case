package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"insider-one-case/internal/idempotency"
	"insider-one-case/internal/model"
	"insider-one-case/internal/repository"
)

var ErrDuplicateEvent = errors.New("event already processed")

type EventService struct {
	repo        *repository.EventRepository
	idempotency *idempotency.RedisStore
	log         *slog.Logger
}

func NewEventService(
	repo *repository.EventRepository,
	idempotency *idempotency.RedisStore,
	log *slog.Logger,
) *EventService {
	return &EventService{
		repo:        repo,
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

	// TODO: enqueue to async worker pipeline and persist batch to ClickHouse.
	if err := s.repo.InsertEvent(ctx, req); err != nil {
		return model.EventIngestResponse{}, err
	}

	return model.EventIngestResponse{
		Status:  "accepted",
		Message: "event queued for processing",
	}, nil
}
