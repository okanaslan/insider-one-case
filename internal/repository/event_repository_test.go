package repository

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"insider-one-case/internal/model"
)

func TestInsertEventsBatchReturnsNilWhenConnIsNil(t *testing.T) {
	repo := NewEventRepository(nil, slog.Default())
	events := []model.EventIngestRequest{
		{
			EventName:  "purchase",
			Channel:    "mobile",
			CampaignID: "cmp_123",
			UserID:     "user_1",
			Timestamp:  1710000000,
			Tags:       []string{"promo"},
			Metadata:   map[string]any{"amount": 100},
		},
	}

	err := repo.InsertEventsBatch(context.Background(), events)
	require.NoError(t, err)
}

func TestInsertEventReturnsNilWhenConnIsNil(t *testing.T) {
	repo := NewEventRepository(nil, slog.Default())
	event := model.EventIngestRequest{
		EventName:  "signup",
		Channel:    "web",
		CampaignID: "cmp_456",
		UserID:     "user_2",
		Timestamp:  1710000100,
		Tags:       []string{"organic"},
	}

	err := repo.InsertEvent(context.Background(), event)
	require.NoError(t, err)
}

func TestSerializeMetadataNilReturnsEmptyObject(t *testing.T) {
	serialized, err := serializeMetadata(nil)
	require.NoError(t, err)
	require.Equal(t, "{}", serialized)
}

func TestSerializeMetadataReturnsValidJSON(t *testing.T) {
	serialized, err := serializeMetadata(map[string]any{
		"amount":   120,
		"currency": "USD",
	})
	require.NoError(t, err)

	decoded := map[string]any{}
	require.NoError(t, json.Unmarshal([]byte(serialized), &decoded))
	require.Equal(t, float64(120), decoded["amount"])
	require.Equal(t, "USD", decoded["currency"])
}
