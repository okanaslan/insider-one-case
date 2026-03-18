package validator

import (
	"context"

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
	return v.validator.Struct(event)
}
