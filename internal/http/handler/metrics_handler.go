package handler

import (
	"net/http"
	"strconv"

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
	eventName := c.Query("event_name")
	if eventName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "event_name is required",
		})
		return
	}

	fromStr := c.Query("from")
	if fromStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "from is required",
		})
		return
	}
	from, err := strconv.ParseInt(fromStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "from must be a unix timestamp",
		})
		return
	}

	toStr := c.Query("to")
	if toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "to is required",
		})
		return
	}
	to, err := strconv.ParseInt(toStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "to must be a unix timestamp",
		})
		return
	}

	if from >= to {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "from must be less than to",
		})
		return
	}

	groupBy := c.Query("group_by")
	if groupBy != "" && groupBy != "channel" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "group_by must be one of: channel",
		})
		return
	}

	query := model.MetricsQuery{EventName: eventName, From: from, To: to, GroupBy: groupBy}

	resp, err := h.metricsService.Query(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to query metrics",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
