package validator

import (
	"context"
	"fmt"
	"strings"

	govalidator "github.com/go-playground/validator/v10"

	"insider-one-case/internal/model"
)

type EventValidator struct {
	validator *govalidator.Validate
}

func NewEventValidator() *EventValidator {
	return &EventValidator{validator: govalidator.New()}
}

func (v *EventValidator) ValidateEvent(ctx context.Context, event model.EventIngestRequest) error {
	_ = ctx

	if strings.TrimSpace(event.EventName) == "" {
		return fmt.Errorf("event_name is required")
	}
	if strings.TrimSpace(event.Channel) == "" {
		return fmt.Errorf("channel is required")
	}
	if strings.TrimSpace(event.CampaignID) == "" {
		return fmt.Errorf("campaign_id is required")
	}
	if strings.TrimSpace(event.UserID) == "" {
		return fmt.Errorf("user_id is required")
	}
	if event.Timestamp <= 0 {
		return fmt.Errorf("timestamp is required and must be positive")
	}
	if len(event.Tags) == 0 {
		return fmt.Errorf("tags is required and must not be empty")
	}

	return v.validator.Struct(event)
}
