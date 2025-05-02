package handler_test

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sendurangr/url-analyzer-api/internal/handler"
	"github.com/sendurangr/url-analyzer-api/internal/model"
	_ "github.com/sendurangr/url-analyzer-api/internal/services"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockAnalyzerService struct {
	shouldFail bool
}

func (m *mockAnalyzerService) AnalyzePage(url string) (*model.AnalyzerResult, error) {
	if m.shouldFail {
		return nil, errors.New("analyze error")
	}
	return &model.AnalyzerResult{HTMLVersion: "HTML5"}, nil
}

func setupRouter(h *handler.AnalyzerHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/url-analyzer", h.UrlAnalyzerHandler)
	return r
}

func TestUrlAnalyzerHandler_MissingURL(t *testing.T) {
	h := handler.NewAnalyzerHandler(&mockAnalyzerService{})
	r := setupRouter(h)

	req, _ := http.NewRequest(http.MethodGet, "/url-analyzer", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "Missing 'url'") {
		t.Errorf("Expected 400 with message about missing URL, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUrlAnalyzerHandler_InvalidURL(t *testing.T) {
	h := handler.NewAnalyzerHandler(&mockAnalyzerService{})
	r := setupRouter(h)

	req, _ := http.NewRequest(http.MethodGet, "/url-analyzer?url=ht!tp://bad_url", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "Invalid or unsupported URL") {
		t.Errorf("Expected 400 for invalid URL, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUrlAnalyzerHandler_AnalyzeError(t *testing.T) {
	h := handler.NewAnalyzerHandler(&mockAnalyzerService{shouldFail: true})
	r := setupRouter(h)

	req, _ := http.NewRequest(http.MethodGet, "/url-analyzer?url=https://valid.com", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), "analyze error") {
		t.Errorf("Expected 500 with analyze error, got %d: %s", w.Code, w.Body.String())
	}
}
