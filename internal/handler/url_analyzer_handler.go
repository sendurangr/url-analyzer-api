package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sendurangr/url-analyzer-api/internal/services"
	"github.com/sendurangr/url-analyzer-api/internal/utils"
	"log/slog"
	"net/http"
)

func UrlAnalyzerHandler(ctx *gin.Context) {
	url := ctx.Query("url")
	if url == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Missing 'url' query parameter")
		return
	}

	response, err := services.AnalyzePage(url)

	if err != nil {
		slog.Error("failed to analyze page",
			"url", url,
			"error", err,
		)
		utils.RespondWithError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, response)
}
