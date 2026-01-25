package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"log/slog"
)

func Logging(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		attrs := []slog.Attr{
			slog.Int("status", statusCode),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("ip", c.ClientIP()),
			slog.Float64("latency_ms", float64(latency.Microseconds())/1000.0),
			slog.String("user_agent", c.Request.UserAgent()),
		}

		if reqID := c.GetHeader("X-Request-ID"); reqID != "" {
			attrs = append(attrs, slog.String("request_id", reqID))
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("gin_errors", c.Errors.String()))
		}

		level := slog.LevelInfo
		if statusCode >= http.StatusInternalServerError {
			level = slog.LevelError
		} else if statusCode >= http.StatusBadRequest {
			level = slog.LevelWarn
		}

		logger.LogAttrs(c.Request.Context(), level, "http_request", attrs...)
	}
}
