package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// @Summary      Teapot
// @Description  I'm a teapot
// @Tags         Miscellaneous
// @Produce      json
// @Success      418
// @Router       /teapot [get]
func TeapotHandler(context *gin.Context) {
	context.Status(http.StatusTeapot)
}
