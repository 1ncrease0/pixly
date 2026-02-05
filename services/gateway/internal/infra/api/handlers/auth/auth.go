package auth

import (
	"context"
	"github.com/1ncrease0/pixly/services/gateway/internal/domain/auth"
	"net/http"
	"time"

	"github.com/1ncrease0/pixly/services/gateway/internal/infra/api/apierr"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/api/middleware"
	"github.com/gin-gonic/gin"
	"log/slog"
)

const refreshTokenCookieName = "refresh_token"
const refreshCookiePath = "/api/v1/auth"

type AuthClient interface {
	Register(ctx context.Context, email, username, password string) error
	VerifyEmail(ctx context.Context, code string) error
	Login(ctx context.Context, email, password string) (*auth.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*auth.TokenPair, error)
	ResendVerification(ctx context.Context, email string) error
	GetUser(ctx context.Context, userID string) (*auth.User, error)
}

type Handler struct {
	log        *slog.Logger
	authClient AuthClient
	accessTTL  time.Duration
	refreshTTL time.Duration
	secure     bool
}

func NewHandler(log *slog.Logger, authClient AuthClient, accessTTL, refreshTTL time.Duration, secure bool) *Handler {
	return &Handler{
		log:        log,
		authClient: authClient,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		secure:     secure,
	}
}

func (h *Handler) setRefreshCookie(c *gin.Context, value string, maxAge int) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		refreshTokenCookieName,
		value,
		maxAge,
		refreshCookiePath,
		"",
		h.secure,
		true,
	)
}

type RegisterRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Username string `json:"username" example:"johndoe"`
	Password string `json:"password" example:"securePassword123"`
}

type RegisterResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Registration successful, verification email sent"`
}

// Register godoc
// @Summary      Register new user
// @Description  Register a new user. A verification email is sent to the given address.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body RegisterRequest true "Registration data"
// @Success      200 {object} RegisterResponse
// @Failure      400 {object} apierr.ErrorResponse "BAD_REQUEST, INVALID_EMAIL, INVALID_USERNAME, INVALID_PASSWORD"
// @Failure      409 {object} apierr.ErrorResponse "USER_ALREADY_EXISTS, USERNAME_TAKEN"
// @Router       /api/v1/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Info("register: invalid request body", slog.Any("error", err))
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
		return
	}

	h.log.Info("register successful", slog.String("email", req.Email), slog.String("username", req.Username))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Registration successful, verification email sent",
	})
}

type MessageResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Email verified successfully"`
}

// VerifyEmail godoc
// @Summary      Verify email
// @Description  Confirm email using the token from the verification email (query param token).
// @Tags         auth
// @Produce      json
// @Param        token query string true "Verification token from email"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} apierr.ErrorResponse "INVALID_VERIFICATION_CODE"
// @Router       /api/v1/auth/verify [post]
func (h *Handler) VerifyEmail(c *gin.Context) {
	code := c.Query("token")
	if code == "" {
		h.log.Info("verify email: missing token")
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

	h.log.Info("verify email successful")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Email verified successfully",
	})
}

type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"password123"`
}

type TokenPairResponse struct {
	AccessToken   string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	AccessTTLSec  int64  `json:"access_ttl_sec" example:"900"`
	RefreshTTLSec int64  `json:"refresh_ttl_sec" example:"2592000"`
}

// Login godoc
// @Summary      Login
// @Description  Authenticate by email and password. Returns access_token and TTLs. Refresh token is set in httpOnly cookie.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body LoginRequest true "Email and password"
// @Success      200 {object} TokenPairResponse
// @Failure      400 {object} apierr.ErrorResponse
// @Failure      401 {object} apierr.ErrorResponse "UNAUTHENTICATED"
// @Failure      412 {object} apierr.ErrorResponse "USER_NOT_VERIFIED"
// @Router       /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Info("login: invalid request body", slog.Any("error", err))
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
		resp := apierr.ErrorToResponse(err)
		c.JSON(resp.Status, resp)
		return
	}

	h.setRefreshCookie(c, tokens.RefreshToken, int(h.refreshTTL.Seconds()))
	h.log.Info("login successful", slog.String("email", req.Email))
	c.JSON(http.StatusOK, gin.H{
		"access_token":    tokens.AccessToken,
		"access_ttl_sec":  int64(h.accessTTL.Seconds()),
		"refresh_ttl_sec": int64(h.refreshTTL.Seconds()),
	})
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token,omitempty" example:""`
}

// Refresh godoc
// @Summary      Refresh tokens
// @Description  Issue new access token using refresh token from httpOnly cookie (or optional body). New refresh token is set in cookie. Returns access_token and TTLs.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body RefreshRequest false "Refresh token (optional if sent via cookie)"
// @Success      200 {object} TokenPairResponse
// @Failure      400 {object} apierr.ErrorResponse "Missing refresh token"
// @Failure      401 {object} apierr.ErrorResponse "SESSION_NOT_FOUND, SESSION_EXPIRED"
// @Router       /api/v1/auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	refreshToken, _ := c.Cookie(refreshTokenCookieName)
	if refreshToken == "" {
		var req RefreshRequest
		_ = c.ShouldBindJSON(&req)
		refreshToken = req.RefreshToken
	}
	if refreshToken == "" {
		h.log.Info("refresh: missing refresh token")
		resp := apierr.NewErrorResponse(
			http.StatusBadRequest,
			apierr.BadRequest,
			"Missing refresh token",
			"refresh_token",
		)
		c.JSON(resp.Status, resp)
		return
	}

	tokens, err := h.authClient.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		resp := apierr.ErrorToResponse(err)
		c.JSON(resp.Status, resp)
		return
	}

	h.setRefreshCookie(c, tokens.RefreshToken, int(h.refreshTTL.Seconds()))
	h.log.Info("refresh successful")
	c.JSON(http.StatusOK, gin.H{
		"access_token":    tokens.AccessToken,
		"access_ttl_sec":  int64(h.accessTTL.Seconds()),
		"refresh_ttl_sec": int64(h.refreshTTL.Seconds()),
	})
}

type ResendVerificationRequest struct {
	Email string `json:"email" example:"user@example.com"`
}

// ResendVerification godoc
// @Summary      Resend verification email
// @Description  Resend the verification email to the given address.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body ResendVerificationRequest true "User email"
// @Success      200 {object} MessageResponse
// @Failure      400 {object} apierr.ErrorResponse
// @Failure      404 {object} apierr.ErrorResponse "USER_NOT_FOUND"
// @Router       /api/v1/auth/resend-verification [post]
func (h *Handler) ResendVerification(c *gin.Context) {
	var req ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Info("resend verification: invalid request body", slog.Any("error", err))
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

type MeResponse struct {
	ID       string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email    string `json:"email" example:"user@example.com"`
	Username string `json:"username" example:"johndoe"`
}

// Me godoc
// @Summary      Current user
// @Description  Return the authenticated user. Requires Authorization: Bearer &lt;access_token&gt;.
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} MeResponse
// @Failure      401 {object} apierr.ErrorResponse "UNAUTHENTICATED"
// @Router       /api/v1/me [get]
func (h *Handler) Me(c *gin.Context) {
	userIDVal, ok := c.Get(middleware.UserIDKey)
	if !ok {
		h.log.Info("me: userID not found in context")
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
		h.log.Info("me: empty userID")
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
