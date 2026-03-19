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

// GetMetrics returns aggregated metrics for a given event and time range.
// @Summary Query event metrics
// @Description Returns totals and optional channel grouping for an event in the provided [from, to) unix timestamp range.
// @Tags metrics
// @Produce json
// @Param event_name query string true "Event name" example(purchase)
// @Param from query int true "Inclusive unix timestamp lower bound" example(1710000000)
// @Param to query int true "Exclusive unix timestamp upper bound" example(1710086400)
// @Param group_by query string false "Optional grouping" Enums(channel)
// @Success 200 {object} model.MetricsResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /metrics [get]
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	eventName := c.Query("event_name")
	if eventName == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid_request", Message: "event_name is required"})
		return
	}

	fromStr := c.Query("from")
	if fromStr == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid_request", Message: "from is required"})
		return
	}
	from, err := strconv.ParseInt(fromStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid_request", Message: "from must be a unix timestamp"})
		return
	}

	toStr := c.Query("to")
	if toStr == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid_request", Message: "to is required"})
		return
	}
	to, err := strconv.ParseInt(toStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid_request", Message: "to must be a unix timestamp"})
		return
	}

	if from >= to {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid_request", Message: "from must be less than to"})
		return
	}

	groupBy := c.Query("group_by")
	if groupBy != "" && groupBy != "channel" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid_request", Message: "group_by must be one of: channel"})
		return
	}

	query := model.MetricsQuery{EventName: eventName, From: from, To: to, GroupBy: groupBy}

	resp, err := h.metricsService.Query(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "internal_error", Message: "failed to query metrics"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
