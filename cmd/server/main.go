package main

import (
	"github.com/sendurangr/url-analyzer-api/internal/handler"
	"github.com/sendurangr/url-analyzer-api/internal/middleware"
	"github.com/sendurangr/url-analyzer-api/internal/routes"
	"github.com/sendurangr/url-analyzer-api/internal/urlanalyzer"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	r := gin.Default()
	r.Use(middleware.Cors())

	apiGroup := r.Group("/api/v1")
	analyzerHandler := handler.NewAnalyzerHandler(urlanalyzer.NewAnalyzerService())
	routes.SetupRouters(apiGroup, analyzerHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return r.Run(":" + port)
}
