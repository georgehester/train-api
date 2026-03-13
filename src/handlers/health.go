package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary      Health
// @Description  Check if the service is active
// @Tags         Miscellaneous
// @Produce      json
// @Success      200
// @Router       /health [get]
func HealthHandler(context *gin.Context) {
	context.Status(http.StatusOK)
}
