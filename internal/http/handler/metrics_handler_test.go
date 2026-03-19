package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"insider-one-case/internal/model"
	"insider-one-case/internal/service"
)

func TestMetricsHandlerMissingEventNameReturns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &MetricsHandler{metricsService: nil}
	r.GET("/metrics", h.GetMetrics)

	req := httptest.NewRequest(http.MethodGet, "/metrics?from=1710000000&to=1710086400", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "event_name is required")
}

func TestMetricsHandlerInvalidGroupByReturns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &MetricsHandler{metricsService: nil}
	r.GET("/metrics", h.GetMetrics)

	req := httptest.NewRequest(http.MethodGet, "/metrics?event_name=purchase&from=1710000000&to=1710086400&group_by=hour", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "group_by must be one of: channel")
}

func TestMetricsHandlerMissingFromReturns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &MetricsHandler{metricsService: nil}
	r.GET("/metrics", h.GetMetrics)

	req := httptest.NewRequest(http.MethodGet, "/metrics?event_name=purchase&to=1710086400", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "from is required")
}

func TestMetricsHandlerMissingToReturns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &MetricsHandler{metricsService: nil}
	r.GET("/metrics", h.GetMetrics)

	req := httptest.NewRequest(http.MethodGet, "/metrics?event_name=purchase&from=1710000000", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "to is required")
}

func TestMetricsHandlerValidQueryReturns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	repo := &fakeMetricsRepoForHandler{}
	svc := service.NewMetricsService(repo, nil)
	h := NewMetricsHandler(svc)
	r.GET("/metrics", h.GetMetrics)

	req := httptest.NewRequest(http.MethodGet, "/metrics?event_name=purchase&from=1710000000&to=1710086400&group_by=channel", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "total_count")
	require.Contains(t, w.Body.String(), "group_by")
}

type fakeMetricsRepoForHandler struct{}

func (f *fakeMetricsRepoForHandler) QueryTotals(ctx context.Context, query model.MetricsQuery) (uint64, uint64, error) {
	_ = ctx
	_ = query
	return 100, 50, nil
}

func (f *fakeMetricsRepoForHandler) QueryGroupedByChannel(ctx context.Context, query model.MetricsQuery) ([]model.MetricsGroup, error) {
	_ = ctx
	_ = query
	return []model.MetricsGroup{{Key: "mobile", Count: 60, UniqueUsers: 30}}, nil
}
