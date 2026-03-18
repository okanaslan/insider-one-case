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
		EventName: "purchase_completed",
		UserID:    "user-1",
		Channel:   "mobile",
		Timestamp: time.Now().UTC(),
	}

	require.Equal(t, "purchase_completed", req.EventName)
}
