package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleHealthcheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
