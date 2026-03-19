package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"insider-one-case/internal/config"
	"insider-one-case/internal/http/handler"
	"insider-one-case/internal/http/middleware"
)

func Build(
	cfg config.Config,
	log *slog.Logger,
	healthHandler *handler.HealthHandler,
	eventHandler *handler.EventHandler,
	metricsHandler *handler.MetricsHandler,
) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(middleware.Recovery(log))
	r.Use(middleware.RequestID())
	r.Use(middleware.Logging(log))

	r.GET("/health", healthHandler.GetHealth)
	r.POST("/events", eventHandler.PostEvent)
	r.POST("/events/bulk", eventHandler.PostEventBulk)
	r.GET("/metrics", metricsHandler.GetMetrics)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
