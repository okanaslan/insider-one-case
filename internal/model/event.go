package model

import "time"

type EventIngestRequest struct {
	EventID   string         `json:"event_id"`
	EventName string         `json:"event_name" validate:"required"`
	UserID    string         `json:"user_id" validate:"required"`
	Channel   string         `json:"channel" validate:"required"`
	Timestamp time.Time      `json:"timestamp" validate:"required"`
	Payload   map[string]any `json:"payload"`
}

type EventIngestResponse struct {
	EventID string `json:"event_id"`
	Status  string `json:"status"`
}
