package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sendurangr/url-analyzer-api/internal/constants"
	"github.com/sendurangr/url-analyzer-api/internal/handler"
	"github.com/sendurangr/url-analyzer-api/internal/middleware"
	"github.com/sendurangr/url-analyzer-api/internal/routes"
	"github.com/sendurangr/url-analyzer-api/internal/urlanalyzer"
	"log/slog"
	"net/http"
	"os"
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

	r.GET("/health", handler.HealthCheckHandler)

	httpClient := &http.Client{
		Timeout: constants.HttpClientTimeout,
	}

	apiGroup := r.Group("/api/v1")
	analyzerHandler := handler.NewAnalyzerHandler(urlanalyzer.NewAnalyzerService(httpClient))
	routes.SetupRouters(apiGroup, analyzerHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return r.Run(":" + port)
}
