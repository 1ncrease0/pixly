package art

import (
	"context"

	"log/slog"
	"net/http"

	artdomain "github.com/1ncrease0/pixly/services/gateway/internal/domain/art"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/api/apierr"
	"github.com/1ncrease0/pixly/services/gateway/internal/infra/api/middleware"
	"github.com/gin-gonic/gin"
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

// GetUserPreviews godoc
// @Summary      Get user pixelarts previews
// @Description  Get list of all pixelart previews for the authenticated user. Requires Authorization: Bearer &lt;access_token&gt;.
// @Tags         art
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} GetUserPreviewsResponse
// @Failure      401 {object} apierr.ErrorResponse "UNAUTHENTICATED"
// @Failure      500 {object} apierr.ErrorResponse "INTERNAL_ERROR"
// @Router       /api/v1/pixelarts [get]
func (h *Handler) GetUserPreviews(c *gin.Context) {
	userIDVal, ok := c.Get(middleware.UserIDKey)
	if !ok {
		h.log.Info("get user previews: userID not found in context")
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

	in := artdomain.GetUserPreviewsInput{
		UserID: userID,
	}

	previews, err := h.artClient.GetUserPreviews(c.Request.Context(), in)
	if err != nil {
		resp := apierr.ErrorToResponse(err)
		c.JSON(resp.Status, resp)
		return
	}

	out := make([]PreviewResponse, len(previews))
	for i, p := range previews {
		out[i] = PreviewResponse{
			PixelartID: p.PixelartID,
			Title:      p.Title,
			ImageURL:   p.ImageURL,
		}
	}

	h.log.Debug("get user previews successful", slog.String("user_id", userID), slog.Int("count", len(previews)))
	c.JSON(http.StatusOK, GetUserPreviewsResponse{Previews: out})
}

type GetUserPreviewsResponse struct {
	Previews []PreviewResponse `json:"previews"`
}

type PreviewResponse struct {
	PixelartID string `json:"pixelart_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title      string `json:"title" example:"My pixel art"`
	ImageURL   string `json:"image_url" example:"https://minio.example.com/pixelarts/abc.png"`
}

// GetUserPixelart godoc
// @Summary      Get pixelart by ID
// @Description  Get a specific pixelart by ID for the authenticated user. Requires Authorization: Bearer &lt;access_token&gt;.
// @Tags         art
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Pixelart ID"
// @Success      200 {object} GetUserPixelartResponse
// @Failure      400 {object} apierr.ErrorResponse "BAD_REQUEST"
// @Failure      401 {object} apierr.ErrorResponse "UNAUTHENTICATED"
// @Failure      404 {object} apierr.ErrorResponse "PIXELART_NOT_FOUND"
// @Router       /api/v1/pixelart/{id} [get]
func (h *Handler) GetUserPixelart(c *gin.Context) {
	userIDVal, ok := c.Get(middleware.UserIDKey)
	if !ok {
		h.log.Info("get user pixelart: userID not found in context")
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

	pixelartID := c.Param("id")
	if pixelartID == "" {
		resp := apierr.NewErrorResponse(
			http.StatusBadRequest,
			apierr.BadRequest,
			"pixelart ID is required",
			"id",
		)
		c.JSON(resp.Status, resp)
		return
	}

	in := artdomain.GetUserPixelartInput{
		UserID:     userID,
		PixelartID: pixelartID,
	}

	art, err := h.artClient.GetUserPixelart(c.Request.Context(), in)
	if err != nil {
		resp := apierr.ErrorToResponse(err)
		c.JSON(resp.Status, resp)
		return
	}

	h.log.Debug("get user pixelart successful", slog.String("user_id", userID), slog.String("pixelart_id", pixelartID))
	c.JSON(http.StatusOK, GetUserPixelartResponse{
		Title:    art.Title,
		Palette:  art.Palette,
		Pixels:   art.Pixels,
		Width:    art.Width,
		Height:   art.Height,
		ImageURL: art.ImageURL,
	})
}

type GetUserPixelartResponse struct {
	Title    string   `json:"title" example:"My pixel art"`
	Palette  []string `json:"palette" example:"#000000,#ffffff"`
	Pixels   []int64  `json:"pixels"`
	Width    int64    `json:"width" example:"16"`
	Height   int64    `json:"height" example:"16"`
	ImageURL string   `json:"image_url" example:"https://minio.example.com/pixelarts/abc.png"`
}

type UpdateCanvasRequest struct {
	Palette []string `json:"palette" binding:"required" example:"#000000,#ffffff"`
	Pixels  []int64  `json:"pixels" binding:"required"`
	Width   int64    `json:"width" binding:"required,gt=0" example:"16"`
	Height  int64    `json:"height" binding:"required,gt=0" example:"16"`
}

// UpdateCanvas godoc
// @Summary      Update pixelart canvas
// @Description  Update the canvas of an existing pixelart. Requires Authorization: Bearer &lt;access_token&gt;.
// @Tags         art
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Pixelart ID"
// @Param        body body UpdateCanvasRequest true "Canvas data (palette, pixels, width, height)"
// @Success      200 {object} map[string]string "success"
// @Failure      400 {object} apierr.ErrorResponse "BAD_REQUEST"
// @Failure      401 {object} apierr.ErrorResponse "UNAUTHENTICATED"
// @Failure      404 {object} apierr.ErrorResponse "PIXELART_NOT_FOUND"
// @Router       /api/v1/pixelart/{id} [patch]
func (h *Handler) UpdateCanvas(c *gin.Context) {
	userIDVal, ok := c.Get(middleware.UserIDKey)
	if !ok {
		h.log.Info("update canvas: userID not found in context")
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

	pixelartID := c.Param("id")
	if pixelartID == "" {
		resp := apierr.NewErrorResponse(
			http.StatusBadRequest,
			apierr.BadRequest,
			"pixelart ID is required",
			"id",
		)
		c.JSON(resp.Status, resp)
		return
	}

	var req UpdateCanvasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Info("update canvas: invalid request body", slog.Any("error", err))
		resp := apierr.NewErrorResponse(
			http.StatusBadRequest,
			apierr.BadRequest,
			"Invalid request body",
			"request",
		)
		c.JSON(resp.Status, resp)
		return
	}

	in := artdomain.UpdateCanvasInput{
		UserID:     userID,
		PixelartID: pixelartID,
		Palette:    req.Palette,
		Pixels:     req.Pixels,
		Width:      req.Width,
		Height:     req.Height,
	}

	if err := h.artClient.UpdateCanvas(c.Request.Context(), in); err != nil {
		resp := apierr.ErrorToResponse(err)
		c.JSON(resp.Status, resp)
		return
	}

	h.log.Info("update canvas successful", slog.String("user_id", userID), slog.String("pixelart_id", pixelartID))
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// DeletePixelart godoc
// @Summary      Delete pixelart
// @Description  Delete a pixelart by ID. Requires Authorization: Bearer &lt;access_token&gt;.
// @Tags         art
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Pixelart ID"
// @Success      200 {object} map[string]string "success"
// @Failure      400 {object} apierr.ErrorResponse "BAD_REQUEST"
// @Failure      401 {object} apierr.ErrorResponse "UNAUTHENTICATED"
// @Failure      404 {object} apierr.ErrorResponse "PIXELART_NOT_FOUND"
// @Router       /api/v1/pixelart/{id} [delete]
func (h *Handler) DeletePixelart(c *gin.Context) {
	userIDVal, ok := c.Get(middleware.UserIDKey)
	if !ok {
		h.log.Info("delete pixelart: userID not found in context")
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

	pixelartID := c.Param("id")
	if pixelartID == "" {
		resp := apierr.NewErrorResponse(
			http.StatusBadRequest,
			apierr.BadRequest,
			"pixelart ID is required",
			"id",
		)
		c.JSON(resp.Status, resp)
		return
	}

	in := artdomain.DeletePixelartInput{
		UserID:     userID,
		PixelartID: pixelartID,
	}

	if err := h.artClient.DeletePixelart(c.Request.Context(), in); err != nil {
		resp := apierr.ErrorToResponse(err)
		c.JSON(resp.Status, resp)
		return
	}

	h.log.Info("delete pixelart successful", slog.String("user_id", userID), slog.String("pixelart_id", pixelartID))
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
