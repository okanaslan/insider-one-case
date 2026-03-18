package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"insider-one-case/internal/model"
)

func TestErrDuplicateEventDefined(t *testing.T) {
	require.Error(t, ErrDuplicateEvent)
}

func TestEventRequestModelShape(t *testing.T) {
	req := model.EventIngestRequest{
		EventName:  "purchase_completed",
		UserID:     "user-1",
		Channel:    "mobile",
		CampaignID: "cmp_123",
		Timestamp:  time.Now().Unix(),
		Tags:       []string{"promo", "summer"},
	}

	require.Equal(t, "purchase_completed", req.EventName)
	require.Equal(t, "user-1", req.UserID)
}

func TestEventUniquenessKey(t *testing.T) {
	req := model.EventIngestRequest{
		EventName:  "purchase_completed",
		UserID:     "user-1",
		Timestamp:  1710000000,
		Channel:    "mobile",
		CampaignID: "cmp_123",
		Tags:       []string{"tag1"},
	}

	key := req.UniquenessKey()
	require.Equal(t, "user-1|1710000000|purchase_completed", key)
}
