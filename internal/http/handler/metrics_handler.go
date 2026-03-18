package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"insider-one-case/internal/model"
	"insider-one-case/internal/service"
)

type MetricsHandler struct {
	metricsService *service.MetricsService
}

func NewMetricsHandler(metricsService *service.MetricsService) *MetricsHandler {
	return &MetricsHandler{metricsService: metricsService}
}

func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	query := model.MetricsQuery{
		MetricName:  c.DefaultQuery("metric_name", "events_total"),
		Granularity: c.DefaultQuery("granularity", "1m"),
		Limit:       100,
	}

	if fromStr := c.Query("from"); fromStr != "" {
		if parsed, err := time.Parse(time.RFC3339, fromStr); err == nil {
			query.From = parsed
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if parsed, err := time.Parse(time.RFC3339, toStr); err == nil {
			query.To = parsed
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			query.Limit = parsed
		}
	}

	resp, err := h.metricsService.Query(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "failed to query metrics",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    resp,
	})
}
