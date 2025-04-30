package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sendurangr/url-analyzer-api/internal/services"
	"log/slog"
	"net/http"
)

func UrlAnalyzerHandler(ctx *gin.Context) {
	url := ctx.Query("url")
	if url == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'url' query parameter"})
		return
	}

	response, err := services.AnalyzePage(url)
	if err != nil {
		slog.Error("AnalyzePage error:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}
