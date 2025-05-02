package main

import (
	"github.com/sendurangr/url-analyzer-api/internal/handler"
	"github.com/sendurangr/url-analyzer-api/internal/middleware"
	"github.com/sendurangr/url-analyzer-api/internal/routes"
	"github.com/sendurangr/url-analyzer-api/internal/services"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(middleware.Cors())

	apiGroup := r.Group("/api/v1")

	analyzerService := services.NewAnalyzerService()
	analyzerHandler := handler.NewAnalyzerHandler(analyzerService)

	routes.SetupRouters(apiGroup, analyzerHandler)

	err := r.Run(":8080")

	if err != nil {
		slog.Error("Server startup failed", "error", err)
		os.Exit(1)
	}
}
