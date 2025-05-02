package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sendurangr/url-analyzer-api/internal/services"
	"github.com/sendurangr/url-analyzer-api/internal/utils"
	"log/slog"
	"net/http"
	"net/url"
)

type AnalyzerHandler struct {
	Service services.AnalyzerService
}

func NewAnalyzerHandler(svc services.AnalyzerService) *AnalyzerHandler {
	return &AnalyzerHandler{Service: svc}
}

func (h *AnalyzerHandler) UrlAnalyzerHandler(ctx *gin.Context) {
	rawURL := ctx.Query("url")
	if rawURL == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Missing 'url' query parameter")
		return
	}

	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		utils.RespondWithError(ctx, http.StatusBadRequest, "Invalid or unsupported URL scheme")
		return
	}

	response, err := h.Service.AnalyzePage(rawURL)
	if err != nil {
		slog.Error("Failed to analyze page", "url", rawURL, "error", err)
		utils.RespondWithError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, response)
}
