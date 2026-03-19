package repository

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"insider-one-case/internal/model"
)

func TestQueryTotalsReturnsEmptyWhenConnIsNil(t *testing.T) {
	repo := NewMetricsRepository(nil, slog.Default())

	total, unique, err := repo.QueryTotals(context.Background(), model.MetricsQuery{
		EventName: "purchase",
		From:      1710000000,
		To:        1710086400,
	})

	require.NoError(t, err)
	require.Equal(t, uint64(0), total)
	require.Equal(t, uint64(0), unique)
}

func TestQueryGroupedByChannelReturnsEmptyWhenConnIsNil(t *testing.T) {
	repo := NewMetricsRepository(nil, slog.Default())

	groups, err := repo.QueryGroupedByChannel(context.Background(), model.MetricsQuery{
		EventName: "purchase",
		From:      1710000000,
		To:        1710086400,
		GroupBy:   "channel",
	})

	require.NoError(t, err)
	require.Len(t, groups, 0)
}
