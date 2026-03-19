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

	queue := &fakeEventEnqueuer{err: errors.New("queue full")}
	svc := service.NewEventService(queue, idempotency.NewRedisStore(nil, slog.Default()), slog.Default())
	h := NewEventHandler(svc, appvalidator.NewEventValidator())

	r.POST("/events", h.PostEvent)

	payload := `{"event_name":"purchase","channel":"mobile","campaign_id":"cmp_1","user_id":"user_1","timestamp":1710000000,"tags":["promo"],"metadata":{"amount":120}}`
	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusTooManyRequests, w.Code)
	require.Contains(t, w.Body.String(), "rate_limited")
}
