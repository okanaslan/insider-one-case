package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"insider-one-case/internal/model"
	"insider-one-case/internal/service"
	appvalidator "insider-one-case/internal/validator"
)

type EventHandler struct {
	eventService   *service.EventService
	eventValidator *appvalidator.EventValidator
}

func NewEventHandler(eventService *service.EventService, eventValidator *appvalidator.EventValidator) *EventHandler {
	return &EventHandler{eventService: eventService, eventValidator: eventValidator}
}

func (h *EventHandler) PostEvent(c *gin.Context) {
	var req model.EventIngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "malformed JSON body",
		})
		return
	}

	if err := h.eventValidator.ValidateEvent(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	resp, err := h.eventService.Ingest(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateEvent) {
			c.JSON(http.StatusConflict, model.APIResponse{
				Success: false,
				Error:   "duplicate event",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "failed to ingest event",
		})
		return
	}

	c.JSON(http.StatusAccepted, model.APIResponse{
		Success: true,
		Message: "event accepted",
		Data:    resp,
	})
}
