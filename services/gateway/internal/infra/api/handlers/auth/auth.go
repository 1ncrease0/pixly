package auth

import (
	"context"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/api/apierr"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

type AuthClient interface {
	Register(ctx context.Context, username, email, password string) error
	VerifyEmail(ctx context.Context, code string) error
}

type Handler struct {
	log        *slog.Logger
	authClient AuthClient
}

func NewHandler(log *slog.Logger, authClient AuthClient) *Handler {
	return &Handler{
		log:        log,
		authClient: authClient,
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp := apierr.NewErrorResponse(
			http.StatusBadRequest,
			apierr.BadRequest,
			"Invalid request body",
			"request",
		)
		c.JSON(resp.Status, resp)
		return
	}

	err := h.authClient.Register(c.Request.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		resp := apierr.ErrorToResponse(err)
		c.JSON(resp.Status, resp)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Registration successful, verification email sent",
	})
}

func (h *Handler) VerifyEmail(c *gin.Context) {
	code := c.Query("token")
	if code == "" {
		resp := apierr.NewErrorResponse(
			http.StatusBadRequest,
			apierr.BadRequest,
			"Missing verification token",
			"token",
		)
		c.JSON(resp.Status, resp)
		return
	}

	err := h.authClient.VerifyEmail(c.Request.Context(), code)
	if err != nil {
		resp := apierr.ErrorToResponse(err)
		c.JSON(resp.Status, resp)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Email verified successfully",
	})
}
