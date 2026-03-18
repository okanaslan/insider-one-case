package model

import "fmt"

type EventIngestRequest struct {
	EventName  string         `json:"event_name" validate:"required"`
	Channel    string         `json:"channel" validate:"required"`
	CampaignID string         `json:"campaign_id" validate:"required"`
	UserID     string         `json:"user_id" validate:"required"`
	Timestamp  int64          `json:"timestamp" validate:"required"`
	Tags       []string       `json:"tags" validate:"required"`
	Metadata   map[string]any `json:"metadata"`
}

// UniquenessKey returns a composite key for deduplication based on user_id, timestamp, and event_name.
func (r *EventIngestRequest) UniquenessKey() string {
	return fmt.Sprintf("%s|%d|%s", r.UserID, r.Timestamp, r.EventName)
}

type EventIngestResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
