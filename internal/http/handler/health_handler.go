package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"insider-one-case/internal/config"
	"insider-one-case/internal/model"
)

type HealthHandler struct {
	cfg config.Config
}

func NewHealthHandler(cfg config.Config) *HealthHandler {
	return &HealthHandler{cfg: cfg}
}

// GetHealth returns service health status.
// @Summary Health check
// @Description Returns basic service metadata and current UTC time.
// @Tags health
// @Produce json
// @Success 200 {object} model.HealthResponse
// @Router /health [get]
func (h *HealthHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, model.HealthResponse{
		Success: true,
		Message: "ok",
		Data: model.HealthData{
			Status: "ok",
			App:    h.cfg.AppName,
			Env:    h.cfg.AppEnv,
			Time:   time.Now().UTC().Format(time.RFC3339),
		},
	})
}
