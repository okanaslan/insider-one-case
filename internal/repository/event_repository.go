package repository

import (
	"context"
	"log/slog"

	"github.com/ClickHouse/clickhouse-go/v2"

	"insider-one-case/internal/model"
)

type EventRepository struct {
	conn clickhouse.Conn
	log  *slog.Logger
}

func NewEventRepository(conn clickhouse.Conn, log *slog.Logger) *EventRepository {
	return &EventRepository{conn: conn, log: log}
}

func (r *EventRepository) InsertEvent(ctx context.Context, event model.EventIngestRequest) error {
	if r.conn == nil {
		r.log.Debug("clickhouse not configured, skipping event insert", "uniqueness_key", event.UniquenessKey())
		return nil
	}

	_ = ctx
	_ = event
	// TODO: implement insert statement and batch flow for event writes.
	return nil
}
