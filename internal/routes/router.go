package routes

import (
	"github.com/sendurangr/url-analyzer-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouters(router *gin.Engine) {
	router.GET("/", handler.HealthCheckHandler)
}
