package art

import (
	"context"

	artdomain "github.com/1ncrease0/pixly/services/gateway/internal/domain/art"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/api/apierr"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/api/middleware"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

type ArtClient interface {
	SavePixelart(ctx context.Context, in artdomain.SavePixelartInput) (string, error)
	UpdateCanvas(ctx context.Context, in artdomain.UpdateCanvasInput) error
	DeletePixelart(ctx context.Context, in artdomain.DeletePixelartInput) error
	GetUserPixelart(ctx context.Context, in artdomain.GetUserPixelartInput) (*artdomain.Pixelart, error)
	GetUserPreviews(ctx context.Context, in artdomain.GetUserPreviewsInput) ([]artdomain.Preview, error)
}

type Handler struct {
	log       *slog.Logger
	artClient ArtClient
}

func NewHandler(log *slog.Logger, artClient ArtClient) *Handler {
	return &Handler{
		log:       log,
		artClient: artClient,
	}
}

type SavePixelartRequest struct {
	Title   string   `json:"title" binding:"required" example:"My pixel art"`
	Palette []string `json:"palette" binding:"required" example:"#000000,#ffffff"`
	Pixels  []int64  `json:"pixels" binding:"required"`
	Width   int64    `json:"width" binding:"required,gt=0" example:"16"`
	Height  int64    `json:"height" binding:"required,gt=0" example:"16"`
}

type SavePixelartResponse struct {
	PixelartID string `json:"pixelart_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// SavePixelart godoc
// @Summary      Save pixelart
// @Description  Create a new pixelart for the authenticated user. Requires Authorization: Bearer &lt;access_token&gt;.
// @Tags         art
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body SavePixelartRequest true "Pixelart data (title, palette, pixels, width, height)"
// @Success      201 {object} SavePixelartResponse
// @Failure      400 {object} apierr.ErrorResponse "BAD_REQUEST, invalid body or validation"
// @Failure      401 {object} apierr.ErrorResponse "UNAUTHENTICATED"
// @Failure      409 {object} apierr.ErrorResponse "PIXELART_CONFLICT"
// @Router       /api/v1/pixelart [post]
func (h *Handler) SavePixelart(c *gin.Context) {
	userIDVal, ok := c.Get(middleware.UserIDKey)
	if !ok {
		h.log.Info("save pixelart: userID not found in context")
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

	var req SavePixelartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Info("save pixelart: invalid request body", slog.Any("error", err))
		resp := apierr.NewErrorResponse(
			http.StatusBadRequest,
			apierr.BadRequest,
			"Invalid request body",
			"request",
		)
		c.JSON(resp.Status, resp)
		return
	}

	in := artdomain.SavePixelartInput{
		UserID:  userID,
		Title:   req.Title,
		Palette: req.Palette,
		Pixels:  req.Pixels,
		Width:   req.Width,
		Height:  req.Height,
	}

	pixelartID, err := h.artClient.SavePixelart(c.Request.Context(), in)
	if err != nil {
		resp := apierr.ErrorToResponse(err)
		c.JSON(resp.Status, resp)
		return
	}

	h.log.Info("save pixelart successful", slog.String("user_id", userID), slog.String("pixelart_id", pixelartID))
	c.JSON(http.StatusCreated, SavePixelartResponse{PixelartID: pixelartID})
}
