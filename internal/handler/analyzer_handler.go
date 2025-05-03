package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sendurangr/url-analyzer-api/internal/urlanalyzer"
	"github.com/sendurangr/url-analyzer-api/internal/utils"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

type AnalyzerHandler struct {
	Service urlanalyzer.AnalyzerService
}

func NewAnalyzerHandler(svc urlanalyzer.AnalyzerService) *AnalyzerHandler {
	return &AnalyzerHandler{Service: svc}
}

func (h *AnalyzerHandler) UrlAnalyzerHandler(ctx *gin.Context) {

	rawURL := ctx.Query("url")
	if rawURL == "" {
		slog.Warn("Missing 'url' query parameter")
		utils.RespondWithError(ctx, http.StatusBadRequest, "Missing 'url' query parameter")
		return
	}

	// Validate the URL format - and not supporting other schemes like ftp or file
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		slog.Warn("Invalid or unsupported URL scheme", "url", rawURL)
		utils.RespondWithError(ctx, http.StatusBadRequest, "Invalid or unsupported URL. Please use http or https.")
		return
	}

	result, err := h.Service.AnalyzePage(rawURL)
	if err != nil {
		slog.Error("Failed to analyze page", "url", rawURL, "error", err)

		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "HTTP error") {
			status = http.StatusBadGateway
		}
		utils.RespondWithError(ctx, status, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, result)
}
