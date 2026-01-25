package api

import (
	"github.com/1ncrease0/pixly/services/gateway/internal/config"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/api/handlers/auth"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/api/handlers/healthcheck"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/api/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log/slog"
)

func InitRoutes(cfg *config.Config, log *slog.Logger, authClient auth.AuthClient) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Logging(log))
	r.Use(gin.Recovery())

	if cfg.HTTP.CORS.Enabled {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     cfg.HTTP.CORS.AllowedOrigins,
			AllowMethods:     cfg.HTTP.CORS.AllowedMethods,
			AllowHeaders:     cfg.HTTP.CORS.AllowedHeaders,
			MaxAge:           cfg.HTTP.CORS.MaxAge,
			AllowCredentials: cfg.HTTP.CORS.AllowCredentials,
		}))
	}

	health := healthcheck.NewHandler()
	auth := auth.NewHandler(log, authClient)

	v1 := r.Group("api/v1", middleware.Logging(log))
	{
		v1.GET("/health", health.Health)
		authentification := v1.Group("/auth")
		{
			authentification.POST("/register", auth.Register)
			authentification.POST("/verify", auth.VerifyEmail)
		}
	}
	return r
}
