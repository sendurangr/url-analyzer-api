package routes

import (
	"github.com/sendurangr/url-analyzer-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouters(router *gin.RouterGroup) {
	router.GET("/health-check", handler.HealthCheckHandler)
	router.GET("/url-analyzer", handler.UrlAnalyzerHandler)
}
