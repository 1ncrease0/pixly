package auth

import (
	"context"
	"github.com/1ncrease0/pixly/services/gateway/internal/domain"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/api/apierr"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/api/middleware"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

type AuthClient interface {
	Register(ctx context.Context, email, username, password string) error
	VerifyEmail(ctx context.Context, code string) error
	Login(ctx context.Context, email, password string) (*domain.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	ResendVerification(ctx context.Context, email string) error
	GetUser(ctx context.Context, userID string) (*domain.User, error)
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
		h.log.Warn("register: invalid request body", slog.Any("error", err))
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
		h.log.Debug("register failed", slog.String("email", req.Email), slog.String("username", req.Username), slog.Any("error", err))
		resp := apierr.ErrorToResponse(err)
		c.JSON(resp.Status, resp)
		return
	}

	h.log.Info("register successful", slog.String("email", req.Email), slog.String("username", req.Username))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Registration successful, verification email sent",
	})
}

func (h *Handler) VerifyEmail(c *gin.Context) {
	code := c.Query("token")
	if code == "" {
		h.log.Warn("verify email: missing token")
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
		h.log.Debug("verify email failed", slog.Any("error", err))
		resp := apierr.ErrorToResponse(err)
		c.JSON(resp.Status, resp)
		return
	}

	h.log.Info("verify email successful")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Email verified successfully",
	})
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("login: invalid request body", slog.Any("error", err))
		resp := apierr.NewErrorResponse(
			http.StatusBadRequest,
			apierr.BadRequest,
			"Invalid request body",
			"request",
		)
		c.JSON(resp.Status, resp)
		return
	}

	tokens, err := h.authClient.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.log.Debug("login failed", slog.String("email", req.Email), slog.Any("error", err))
		resp := apierr.ErrorToResponse(err)
		c.JSON(resp.Status, resp)
		return
	}

	h.log.Info("login successful", slog.String("email", req.Email))
	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("refresh: invalid request body", slog.Any("error", err))
		resp := apierr.NewErrorResponse(
			http.StatusBadRequest,
			apierr.BadRequest,
			"Invalid request body",
			"request",
		)
		c.JSON(resp.Status, resp)
		return
	}

	tokens, err := h.authClient.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.log.Debug("refresh failed", slog.Any("error", err))
		resp := apierr.ErrorToResponse(err)
		c.JSON(resp.Status, resp)
		return
	}

	h.log.Info("refresh successful")
	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

type ResendVerificationRequest struct {
	Email string `json:"email"`
}

func (h *Handler) ResendVerification(c *gin.Context) {
	var req ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Warn("resend verification: invalid request body", slog.Any("error", err))
		resp := apierr.NewErrorResponse(
			http.StatusBadRequest,
			apierr.BadRequest,
			"Invalid request body",
			"request",
		)
		c.JSON(resp.Status, resp)
		return
	}

	err := h.authClient.ResendVerification(c.Request.Context(), req.Email)
	if err != nil {
		h.log.Debug("resend verification failed", slog.String("email", req.Email), slog.Any("error", err))
		resp := apierr.ErrorToResponse(err)
		c.JSON(resp.Status, resp)
		return
	}

	h.log.Info("resend verification successful", slog.String("email", req.Email))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Verification email sent",
	})
}

func (h *Handler) Me(c *gin.Context) {
	userIDVal, ok := c.Get(middleware.UserIDKey)
	if !ok {
		h.log.Warn("me: userID not found in context")
		resp := apierr.NewErrorResponse(
			http.StatusUnauthorized,
			apierr.Unauthenticated,
			"unauthorized",
			"",
		)
		c.JSON(resp.Status, resp)
		return
	}

	userID, _ := userIDVal.(string)
	if userID == "" {
		h.log.Warn("me: empty userID")
		resp := apierr.NewErrorResponse(
			http.StatusUnauthorized,
			apierr.Unauthenticated,
			"unauthorized",
			"",
		)
		c.JSON(resp.Status, resp)
		return
	}

	user, err := h.authClient.GetUser(c.Request.Context(), userID)
	if err != nil {
		h.log.Debug("me: get user failed", slog.String("user_id", userID), slog.Any("error", err))
		resp := apierr.ErrorToResponse(err)
		c.JSON(resp.Status, resp)
		return
	}

	h.log.Debug("me: get user successful", slog.String("user_id", userID))
	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
	})
}
