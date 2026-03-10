package api

import (
	"github.com/gin-gonic/gin"
	"vulpz/train-api/src/model"
)

// Helper function to generate an error message response
func SendErrorResponse(context *gin.Context, statusCode int, message string) {
	context.JSON(statusCode, model.ErrorResponse{
		Message: message,
	})
}
