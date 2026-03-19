package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"insider-one-case/internal/idempotency"
	"insider-one-case/internal/model"
)

type fakeEnqueuer struct {
	err      error
	enqueued []model.EventIngestRequest
}

func (f *fakeEnqueuer) Enqueue(ctx context.Context, event model.EventIngestRequest) error {
	_ = ctx
	if f.err != nil {
		return f.err
	}
	f.enqueued = append(f.enqueued, event)
	return nil
}

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

func TestIngestEnqueuesEventSuccessfully(t *testing.T) {
	queue := &fakeEnqueuer{}
	svc := NewEventService(queue, idempotency.NewRedisStore(nil, slog.Default()), slog.Default())

	req := model.EventIngestRequest{
		EventName:  "purchase_completed",
		UserID:     "user-1",
		Timestamp:  1710000000,
		Channel:    "mobile",
		CampaignID: "cmp_123",
		Tags:       []string{"tag1"},
	}

	resp, err := svc.Ingest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "accepted", resp.Status)
	require.Equal(t, "event accepted for processing", resp.Message)
	require.Len(t, queue.enqueued, 1)
	require.Equal(t, req.EventName, queue.enqueued[0].EventName)
}

func TestIngestReturnsOverloadedWhenQueueReturnsError(t *testing.T) {
	queue := &fakeEnqueuer{err: errors.New("queue full")}
	svc := NewEventService(queue, idempotency.NewRedisStore(nil, slog.Default()), slog.Default())

	req := model.EventIngestRequest{
		EventName:  "purchase_completed",
		UserID:     "user-1",
		Timestamp:  1710000000,
		Channel:    "mobile",
		CampaignID: "cmp_123",
		Tags:       []string{"tag1"},
	}

	_, err := svc.Ingest(context.Background(), req)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrOverloaded)
	require.Len(t, queue.enqueued, 0)
}

func TestIngestBulkAllAccepted(t *testing.T) {
	queue := &fakeEnqueuer{}
	svc := NewEventService(queue, idempotency.NewRedisStore(nil, slog.Default()), slog.Default())

	bulkReq := model.BulkEventIngestRequest{
		Events: []model.EventIngestRequest{
			{
				EventName:  "purchase",
				UserID:     "user_1",
				Timestamp:  1710000000,
				Channel:    "mobile",
				CampaignID: "cmp_1",
				Tags:       []string{"promo"},
			},
			{
				EventName:  "view",
				UserID:     "user_2",
				Timestamp:  1710000001,
				Channel:    "web",
				CampaignID: "cmp_2",
				Tags:       []string{"summer"},
			},
		},
	}

	resp := svc.IngestBulk(context.Background(), bulkReq)
	require.Equal(t, "accepted_all", resp.Status)
	require.Equal(t, 2, resp.Summary.Total)
	require.Equal(t, 2, resp.Summary.Accepted)
	require.Equal(t, 0, resp.Summary.Duplicate)
	require.Equal(t, 0, resp.Summary.Overloaded)
	require.Len(t, queue.enqueued, 2)
}

func TestIngestBulkMixedOutcomes(t *testing.T) {
	queue := &fakeEnqueuer{err: errors.New("queue full")}
	svc := NewEventService(queue, idempotency.NewRedisStore(nil, slog.Default()), slog.Default())

	bulkReq := model.BulkEventIngestRequest{
		Events: []model.EventIngestRequest{
			{
				EventName:  "purchase",
				UserID:     "user_1",
				Timestamp:  1710000000,
				Channel:    "mobile",
				CampaignID: "cmp_1",
				Tags:       []string{"promo"},
			},
			{
				EventName:  "view",
				UserID:     "user_2",
				Timestamp:  1710000001,
				Channel:    "web",
				CampaignID: "cmp_2",
				Tags:       []string{"summer"},
			},
		},
	}

	resp := svc.IngestBulk(context.Background(), bulkReq)
	require.Equal(t, "accepted_partial", resp.Status)
	require.Equal(t, 2, resp.Summary.Total)
	require.Equal(t, 0, resp.Summary.Accepted)
	require.Equal(t, 2, resp.Summary.Error) // Regular errors map to error, not overload
}
