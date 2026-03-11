package handlers

import (
	"github.com/gin-gonic/gin"
)

// @Summary      Get Documentation
// @Description  Returns the OpenAPI JSON documentation
// @Tags         Documentation
// @Produce      json
// @Success      200  {file}  file "Swagger OpenAPI JSON"
// @Router       /documentation [get]
func DocumentationHandler(context *gin.Context) {
	context.File("/documentation/swagger.json")
}
