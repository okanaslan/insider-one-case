package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"insider-one-case/internal/model"
	"insider-one-case/internal/service"
	appvalidator "insider-one-case/internal/validator"
)

type MetricsHandler struct {
	metricsService *service.MetricsService
	validator      *appvalidator.MetricsValidator
}

func NewMetricsHandler(metricsService *service.MetricsService, validator *appvalidator.MetricsValidator) *MetricsHandler {
	return &MetricsHandler{metricsService: metricsService, validator: validator}
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
	var params model.MetricsQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	query, err := h.validator.ParseAndValidateQuery(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	resp, err := h.metricsService.Query(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "internal_error", Message: "failed to query metrics"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
