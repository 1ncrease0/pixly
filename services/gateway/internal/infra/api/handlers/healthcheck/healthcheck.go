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

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}
