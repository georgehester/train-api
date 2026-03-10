package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func HealthHandler(context *gin.Context) {
	context.Status(http.StatusOK)
}
