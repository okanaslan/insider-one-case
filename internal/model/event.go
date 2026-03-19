package model

import "fmt"

type EventIngestRequest struct {
	EventName  string         `json:"event_name" validate:"required" example:"purchase"`
	Channel    string         `json:"channel" validate:"required" example:"mobile"`
	CampaignID string         `json:"campaign_id" validate:"required" example:"cmp_123"`
	UserID     string         `json:"user_id" validate:"required" example:"user_456"`
	Timestamp  int64          `json:"timestamp" validate:"required" example:"1710000000"`
	Tags       []string       `json:"tags" validate:"required" example:"promo,spring"`
	Metadata   map[string]any `json:"metadata" swaggertype:"object"`
}

// UniquenessKey returns a composite key for deduplication based on user_id, timestamp, and event_name.
func (r *EventIngestRequest) UniquenessKey() string {
	return fmt.Sprintf("%s|%d|%s", r.UserID, r.Timestamp, r.EventName)
}

type EventIngestResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// BulkEventIngestRequest is the request envelope for bulk event ingestion.
type BulkEventIngestRequest struct {
	Events []EventIngestRequest `json:"events" validate:"required"`
}

// BulkEventIngestSummary summarizes the outcomes across all events in a bulk request.
type BulkEventIngestSummary struct {
	Total      int `json:"total"`
	Accepted   int `json:"accepted"`
	Duplicate  int `json:"duplicate"`
	Invalid    int `json:"invalid"`
	Overloaded int `json:"overloaded"`
	Error      int `json:"error"`
}

// BulkEventIngestResponse is the response envelope for bulk event ingestion with partial-success semantics.
type BulkEventIngestResponse struct {
	Status  string                 `json:"status"` // accepted_all, accepted_partial, rejected
	Summary BulkEventIngestSummary `json:"summary"`
}
