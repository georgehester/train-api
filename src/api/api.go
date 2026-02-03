package api

import "github.com/gin-gonic/gin"

type ErrorResponse struct {
	Message string `json:"message,omitempty"`
}

func SendErrorResponse(context *gin.Context, statusCode int, message string) {
	context.JSON(statusCode, ErrorResponse{
		Message: message,
	})
}
