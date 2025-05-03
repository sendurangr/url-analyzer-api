package integrationtest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sendurangr/url-analyzer-api/internal/handler"
	"github.com/sendurangr/url-analyzer-api/internal/model"
	"github.com/sendurangr/url-analyzer-api/internal/routes"
	"github.com/sendurangr/url-analyzer-api/internal/urlanalyzer"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	analyzer := handler.NewAnalyzerHandler(urlanalyzer.NewAnalyzerService(&http.Client{}))
	api := r.Group("/api/v1")
	routes.SetupRouters(api, analyzer)

	return r
}

func TestUrlAnalyzer_MockSite(t *testing.T) {
	// Mock HTML page
	htmlContent := `
		<!DOCTYPE html>
		<html>
			<head><title>Test Page</title></head>
			<body>
				<h1>Main</h1>
				<h2>Subheading</h2>
				<a href="/internal">Internal</a>
				<a href="https://external.com">External</a>
				<form><input type="text" name="username"/><input type="password" name="pass"/></form>
			</body>
		</html>`

	// Starting mock HTTP server
	mockSite := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(htmlContent))
	}))
	defer mockSite.Close()

	// Call the real API
	router := setupRouter()

	// Set up the request
	req, _ := http.NewRequest("GET", "/api/v1/url-analyzer?url="+mockSite.URL, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.Code)
	}

	var result model.AnalyzerResult
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	// Validate result fields
	if result.PageTitle != "Test Page" {
		t.Errorf("Expected title 'Test Page', got '%s'", result.PageTitle)
	}
	if result.Headings.H1 != 1 {
		t.Errorf("Expected 1 H1, got %d", result.Headings.H1)
	}
	if result.Headings.H2 != 1 {
		t.Errorf("Expected 1 H2, got %d", result.Headings.H2)
	}
	if result.InternalLinks != 1 {
		t.Errorf("Expected 1 internal link, got %d", result.InternalLinks)
	}
	if result.ExternalLinks != 1 {
		t.Errorf("Expected 1 external link, got %d", result.ExternalLinks)
	}
	if !result.LoginFormDetected {
		t.Error("Expected login form to be detected")
	}
}
