package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sendurangr/url-analyzer-api/internal/model"
	"log/slog"
	"net/http"
)

func UrlAnalyzerHandler(context *gin.Context) {
	url := context.Query("url")
	slog.Info(url)
	context.JSON(http.StatusOK, model.AnalyzerResult{})
}
