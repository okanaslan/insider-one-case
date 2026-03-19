package handler

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"insider-one-case/internal/config"
	"insider-one-case/internal/idempotency"
	"insider-one-case/internal/model"
	"insider-one-case/internal/service"
	appvalidator "insider-one-case/internal/validator"
)

type fakeEventEnqueuer struct {
	err error
}

func (f *fakeEventEnqueuer) Enqueue(ctx context.Context, event model.EventIngestRequest) error {
	_ = ctx
	_ = event
	return f.err
}

func TestPostEventReturns429WhenOverloaded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := config.Config{BulkMaxEventsPerRequest: 500}
	queue := &fakeEventEnqueuer{err: errors.New("queue full")}
	svc := service.NewEventService(queue, idempotency.NewRedisStore(nil, slog.Default()), slog.Default())
	h := NewEventHandler(svc, appvalidator.NewEventValidator(), cfg)

	r.POST("/events", h.PostEvent)

	payload := `{"event_name":"purchase","channel":"mobile","campaign_id":"cmp_1","user_id":"user_1","timestamp":1710000000,"tags":["promo"],"metadata":{"amount":120}}`
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusTooManyRequests, w.Code)
	require.Contains(t, w.Body.String(), "rate_limited")
}

func TestPostEventBulkEmptyArray(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := config.Config{BulkMaxEventsPerRequest: 500}
	queue := &fakeEventEnqueuer{}
	svc := service.NewEventService(queue, idempotency.NewRedisStore(nil, slog.Default()), slog.Default())
	h := NewEventHandler(svc, appvalidator.NewEventValidator(), cfg)

	r.POST("/events/bulk", h.PostEventBulk)

	payload := `{"events":[]}`
	req := httptest.NewRequest(http.MethodPost, "/events/bulk", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "invalid_request")
}

func TestPostEventBulkOversizedArray(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := config.Config{BulkMaxEventsPerRequest: 2}
	queue := &fakeEventEnqueuer{}
	svc := service.NewEventService(queue, idempotency.NewRedisStore(nil, slog.Default()), slog.Default())
	h := NewEventHandler(svc, appvalidator.NewEventValidator(), cfg)

	r.POST("/events/bulk", h.PostEventBulk)

	payload := `{"events":[
		{"event_name":"purchase","channel":"mobile","campaign_id":"cmp_1","user_id":"user_1","timestamp":1710000000,"tags":["promo"]},
		{"event_name":"view","channel":"web","campaign_id":"cmp_2","user_id":"user_2","timestamp":1710000001,"tags":["summer"]},
		{"event_name":"click","channel":"web","campaign_id":"cmp_3","user_id":"user_3","timestamp":1710000002,"tags":["sale"]}
	]}`
	req := httptest.NewRequest(http.MethodPost, "/events/bulk", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "must have between 1 and 2 items")
}

func TestPostEventBulkPartialSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cfg := config.Config{BulkMaxEventsPerRequest: 500}
	queue := &fakeEventEnqueuer{}
	svc := service.NewEventService(queue, idempotency.NewRedisStore(nil, slog.Default()), slog.Default())
	h := NewEventHandler(svc, appvalidator.NewEventValidator(), cfg)

	r.POST("/events/bulk", h.PostEventBulk)

	payload := `{"events":[
		{"event_name":"purchase","channel":"mobile","campaign_id":"cmp_1","user_id":"user_1","timestamp":1710000000,"tags":["promo"]},
		{"event_name":"view","channel":"web","campaign_id":"cmp_2","user_id":"user_2","timestamp":1710000001,"tags":["summer"]}
	]}`
	req := httptest.NewRequest(http.MethodPost, "/events/bulk", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusAccepted, w.Code)
	require.Contains(t, w.Body.String(), "accepted_all")
	require.Contains(t, w.Body.String(), `"total":2`)
	require.Contains(t, w.Body.String(), `"accepted":2`)
}
