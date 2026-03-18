package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"insider-one-case/internal/model"
)

func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("panic recovered", "panic", rec, "path", c.Request.URL.Path)
				c.AbortWithStatusJSON(http.StatusInternalServerError, model.APIResponse{
					Success: false,
					Error:   "internal server error",
				})
			}
		}()

		c.Next()
	}
}
