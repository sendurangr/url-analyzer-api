package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthCheckHandler(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{
		"message": "Server is running successfully!!",
	})
}
