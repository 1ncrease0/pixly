package api

// @title           Pixly Gateway API
// @version         1.0
// @description     Pixly gateway HTTP API.
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"log/slog"

	"github.com/1ncrease0/pixly/pkg/jwt"
	"github.com/1ncrease0/pixly/services/gateway/internal/config"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/api/handlers/art"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/api/handlers/auth"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/api/handlers/healthcheck"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/api/middleware"
	artgrpc "github.com/1ncrease0/pixly/services/gateway/internal/infra/grpc/art"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/1ncrease0/pixly/services/gateway/docs"
)

func InitRoutes(cfg *config.Config, log *slog.Logger, authClient auth.AuthClient, artClient *artgrpc.Client, m *jwt.Manager) *gin.Engine {
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

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	health := healthcheck.NewHandler()
	secure := cfg.Env != "local"
	authH := auth.NewHandler(log, authClient, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL, secure)

	artH := art.NewHandler(log, artClient)
	authMiddleware := middleware.Auth(m, log)
	v1 := r.Group("api/v1", middleware.Logging(log))
	{
		v1.GET("/health", health.Health)
		authentification := v1.Group("/auth")
		{
			authentification.POST("/register", authH.Register)
			authentification.POST("/verify", authH.VerifyEmail)
			authentification.POST("/login", authH.Login)
			authentification.POST("/refresh", authH.Refresh)
			authentification.POST("/resend-verification", authH.ResendVerification)
		}

		v1.GET("/me", authMiddleware, authH.Me)
		v1.POST("/pixelart", authMiddleware, artH.SavePixelart)
		v1.GET("/pixelarts", authMiddleware, artH.GetUserPreviews)
		v1.GET("/pixelart/:id", authMiddleware, artH.GetUserPixelart)
		v1.PUT("/pixelart/:id", authMiddleware, artH.UpdateCanvas)
		v1.DELETE("/pixelart/:id", authMiddleware, artH.DeletePixelart)
	}
	return r
}
