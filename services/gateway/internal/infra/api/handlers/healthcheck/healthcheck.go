package healthcheck

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

type HealthResponse struct {
	Status    string `json:"status" example:"ok"`
	Timestamp string `json:"timestamp" example:"2024-01-01T12:00:00Z"`
}

// Health godoc
// @Summary      Health check
// @Description  Returns service status and current time in RFC3339.
// @Tags         health
// @Produce      json
// @Success      200 {object} HealthResponse
// @Router       /api/v1/health [get]
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}
