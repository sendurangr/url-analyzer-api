package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"log/slog"
)

func HealthCheckHandler(context *gin.Context) {
	slog.Info("healthck is called")
	context.JSON(http.StatusOK, gin.H{
		"message": "Server is running successfully!!",
	})
}
