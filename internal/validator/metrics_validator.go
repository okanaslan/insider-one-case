package validator

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"insider-one-case/internal/model"
)

type MetricsValidator struct{}

func NewMetricsValidator() *MetricsValidator {
	return &MetricsValidator{}
}

func (v *MetricsValidator) ParseAndValidateQuery(ctx context.Context, params model.MetricsQueryParams) (model.MetricsQuery, error) {
	_ = ctx

	eventName := strings.TrimSpace(params.EventName)
	if eventName == "" {
		return model.MetricsQuery{}, fmt.Errorf("event_name is required")
	}

	if params.From == "" {
		return model.MetricsQuery{}, fmt.Errorf("from is required")
	}
	from, err := strconv.ParseInt(params.From, 10, 64)
	if err != nil {
		return model.MetricsQuery{}, fmt.Errorf("from must be a unix timestamp")
	}

	if params.To == "" {
		return model.MetricsQuery{}, fmt.Errorf("to is required")
	}
	to, err := strconv.ParseInt(params.To, 10, 64)
	if err != nil {
		return model.MetricsQuery{}, fmt.Errorf("to must be a unix timestamp")
	}

	if from >= to {
		return model.MetricsQuery{}, fmt.Errorf("from must be less than to")
	}

	groupBy := strings.TrimSpace(params.GroupBy)
	if groupBy != "" && groupBy != "channel" {
		return model.MetricsQuery{}, fmt.Errorf("group_by must be one of: channel")
	}

	return model.MetricsQuery{EventName: eventName, From: from, To: to, GroupBy: groupBy}, nil
}
