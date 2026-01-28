package middleware

import (
	"github.com/1ncrease0/pixly/pkg/jwt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"strings"
)

const UserIDKey = "userID"

func Auth(jwtManager *jwt.Manager, log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Debug("auth middleware: missing Authorization header", slog.String("path", c.Request.URL.Path))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing Authorization header",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			log.Debug("auth middleware: invalid Authorization header format", slog.String("path", c.Request.URL.Path))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid Authorization header",
			})
			return
		}

		token := parts[1]
		info, err := jwtManager.Parse(token)
		if err != nil {
			log.Debug("auth middleware: token validation failed", slog.String("path", c.Request.URL.Path), slog.Any("error", err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		c.Set(UserIDKey, info.UserID.String())
		log.Debug("auth middleware: token validated successfully", slog.String("user_id", info.UserID.String()), slog.String("path", c.Request.URL.Path))

		c.Next()
	}
}
