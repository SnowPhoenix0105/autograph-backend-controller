package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func coffeeHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusTeapot, "I'm a teapot")
}
