package repository

import (
	"context"
	"encoding/json"
	"fmt"
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
	return r.InsertEventsBatch(ctx, []model.EventIngestRequest{event})
}

func (r *EventRepository) InsertEventsBatch(ctx context.Context, events []model.EventIngestRequest) error {
	if r.conn == nil {
		r.log.Debug("clickhouse not configured, skipping batch insert", "count", len(events))
		return nil
	}
	if len(events) == 0 {
		return nil
	}

	batch, err := r.conn.PrepareBatch(ctx, "INSERT INTO events (event_name, channel, campaign_id, user_id, timestamp, tags, metadata)")
	if err != nil {
		return fmt.Errorf("prepare event insert batch: %w", err)
	}

	for _, event := range events {
		metadata, err := serializeMetadata(event.Metadata)
		if err != nil {
			return fmt.Errorf("serialize metadata: %w", err)
		}

		if err := batch.Append(
			event.EventName,
			event.Channel,
			event.CampaignID,
			event.UserID,
			event.Timestamp,
			event.Tags,
			metadata,
		); err != nil {
			return fmt.Errorf("append event to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("send event insert batch: %w", err)
	}

	return nil
}

func serializeMetadata(metadata map[string]any) (string, error) {
	if metadata == nil {
		return "{}", nil
	}

	b, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
