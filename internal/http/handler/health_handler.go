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

func (h *HealthHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "ok",
		Data: gin.H{
			"status": "ok",
			"app":    h.cfg.AppName,
			"env":    h.cfg.AppEnv,
			"time":   time.Now().UTC(),
		},
	})
}
