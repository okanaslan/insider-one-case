package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	govalidator "github.com/go-playground/validator/v10"

	"insider-one-case/internal/model"
	"insider-one-case/internal/service"
)

type EventHandler struct {
	eventService *service.EventService
}

func NewEventHandler(eventService *service.EventService) *EventHandler {
	return &EventHandler{eventService: eventService}
}

func (h *EventHandler) PostEvent(c *gin.Context) {
	var req model.EventIngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid request body",
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

		var validationErrs govalidator.ValidationErrors
		if errors.As(err, &validationErrs) {
			c.JSON(http.StatusBadRequest, model.APIResponse{
				Success: false,
				Error:   "validation failed",
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
