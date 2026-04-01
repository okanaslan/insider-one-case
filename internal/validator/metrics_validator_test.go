package validator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"insider-one-case/internal/model"
)

func TestMetricsValidatorParseAndValidateQuerySuccess(t *testing.T) {
	v := NewMetricsValidator()

	params := model.MetricsQueryParams{EventName: "purchase", From: "1710000000", To: "1710086400", GroupBy: "channel"}
	query, err := v.ParseAndValidateQuery(context.Background(), params)
	require.NoError(t, err)
	require.Equal(t, "purchase", query.EventName)
	require.EqualValues(t, 1710000000, query.From)
	require.EqualValues(t, 1710086400, query.To)
	require.Equal(t, "channel", query.GroupBy)
}

func TestMetricsValidatorParseAndValidateQueryMissingEventName(t *testing.T) {
	v := NewMetricsValidator()

	params := model.MetricsQueryParams{EventName: "", From: "1710000000", To: "1710086400"}
	_, err := v.ParseAndValidateQuery(context.Background(), params)
	require.EqualError(t, err, "event_name is required")
}

func TestMetricsValidatorParseAndValidateQueryInvalidRange(t *testing.T) {
	v := NewMetricsValidator()

	params := model.MetricsQueryParams{EventName: "purchase", From: "1710086400", To: "1710000000"}
	_, err := v.ParseAndValidateQuery(context.Background(), params)
	require.EqualError(t, err, "from must be less than to")
}
