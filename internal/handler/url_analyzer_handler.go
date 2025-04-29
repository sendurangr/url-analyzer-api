package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sendurangr/url-analyzer-api/internal/services"
	"log/slog"
	"net/http"
)

func UrlAnalyzerHandler(context *gin.Context) {
	url := context.Query("url")
	response, err := services.AnalyzePage(url)
	if err != nil {
		slog.Error(err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	context.JSON(http.StatusOK, response)

}
