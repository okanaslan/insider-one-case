package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"insider-one-case/internal/config"
	"insider-one-case/internal/model"
	"insider-one-case/internal/service"
	appvalidator "insider-one-case/internal/validator"
)

type EventHandler struct {
	eventService   *service.EventService
	eventValidator *appvalidator.EventValidator
	cfg            config.Config
}

func NewEventHandler(eventService *service.EventService, eventValidator *appvalidator.EventValidator, cfg config.Config) *EventHandler {
	return &EventHandler{eventService: eventService, eventValidator: eventValidator, cfg: cfg}
}

// PostEvent ingests a single event asynchronously.
// @Summary Ingest a single event
// @Description Validates and enqueues one event for asynchronous processing.
// @Tags events
// @Accept json
// @Produce json
// @Param payload body model.EventIngestRequest true "Event ingestion request"
// @Success 202 {object} model.EventIngestResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 409 {object} model.ErrorResponse
// @Failure 429 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /events [post]
func (h *EventHandler) PostEvent(c *gin.Context) {
	var req model.EventIngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid_request", Message: "malformed JSON body"})
		return
	}

	if err := h.eventValidator.ValidateEvent(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	resp, err := h.eventService.Ingest(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateEvent) {
			c.JSON(http.StatusConflict, model.ErrorResponse{Error: "duplicate_event", Message: "event already processed"})
			return
		}

		if errors.Is(err, service.ErrOverloaded) {
			c.JSON(http.StatusTooManyRequests, model.ErrorResponse{Error: "rate_limited", Message: "ingestion queue is overloaded, try again"})
			return
		}

		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "internal_error", Message: "failed to ingest event"})
		return
	}

	c.JSON(http.StatusAccepted, resp)
}

// PostEventBulk handles bulk event ingestion with per-event validation and partial-success semantics.
// @Summary Ingest events in bulk
// @Description Accepts between 1 and configured max events; returns aggregate outcome with partial-success semantics.
// @Tags events
// @Accept json
// @Produce json
// @Param payload body model.BulkEventIngestRequest true "Bulk event ingestion request"
// @Success 202 {object} model.BulkEventIngestResponse
// @Failure 400 {object} model.ErrorResponse
// @Router /events/bulk [post]
func (h *EventHandler) PostEventBulk(c *gin.Context) {
	var req model.BulkEventIngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid_request", Message: "malformed JSON body"})
		return
	}

	// Validate envelope: non-empty and within limits.
	if len(req.Events) == 0 || len(req.Events) > h.cfg.BulkMaxEventsPerRequest {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid_request", Message: fmt.Sprintf("events array must have between 1 and %d items", h.cfg.BulkMaxEventsPerRequest)})
		return
	}

	// Count validation errors.
	invalidCount := 0
	for _, event := range req.Events {
		if err := h.eventValidator.ValidateEvent(c.Request.Context(), event); err != nil {
			invalidCount++
		}
	}

	// Call service for processing.
	resp := h.eventService.IngestBulk(c.Request.Context(), req)

	// Update summary with validation errors.
	if invalidCount > 0 {
		resp.Summary.Invalid = invalidCount
		resp.Summary.Accepted -= invalidCount
		resp.Status = "accepted_partial"
	}

	c.JSON(http.StatusAccepted, resp)
}
