package urlanalyzer

import (
	"fmt"
	"github.com/sendurangr/url-analyzer-api/internal/constants"
	"github.com/sendurangr/url-analyzer-api/internal/model"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type analyzeTestCase struct {
	name            string
	htmlContent     string
	wantHtmlVersion string
	wantTitle       string
	wantLogin       bool
	wantH1Count     int
	wantH2Count     int
	wantH3Count     int
	wantH4Count     int
	wantH5Count     int
	wantH6Count     int
	wantInternal    int
	wantExternal    int
}

var httpClient = &http.Client{
	Timeout: constants.TimeoutSeconds * time.Second,
}

func startTestServer(htmlContent string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, htmlContent)
	}))
}

func simulateSuccessAndFailServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/fail", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	return httptest.NewServer(mux)
}

func TestAnalyzePage_TableDrivenTestCases(t *testing.T) {
	tests := []analyzeTestCase{
		{
			name: "HTML 1",
			htmlContent: `
				<!DOCTYPE html>
				<html><head><title>First Html</title></head>
				<body>
					<h1>Welcome</h1>
					<h2>Subheading</h2>
					<a href="/home">Internal</a>
					<a href="https://google.com">External</a>
					<form><input type="password"/></form>
				</body></html>
			`,
			wantHtmlVersion: "HTML5",
			wantTitle:       "First Html",
			wantLogin:       false,
			wantH1Count:     1,
			wantH2Count:     1,
			wantInternal:    1,
			wantExternal:    1,
		},
		{
			name: "HTML 2",
			htmlContent: `
				<html><head><title>Login Page</title></head>
				<body>
					<h3>Welcome</h3>
					<h4>Subheading</h4>
					<h5>Subheading</h5>
					<h6>Subheading</h6>
					<a href="/home">Internal</a>
					<a href="https://google.com">External</a>
					<a type="text/html">skippable</a>
					<a href="#">skippable</a>
					<form> 	<input type="email"/> 
							<input type="password"/></form>
				</body>
				</html>
			`,
			wantHtmlVersion: "Older HTML or XHTML",
			wantTitle:       "Login Page",
			wantLogin:       true,
			wantH3Count:     1,
			wantH4Count:     1,
			wantH5Count:     1,
			wantH6Count:     1,
			wantInternal:    1,
			wantExternal:    1,
		},
		{
			name:            "Page with no headings or links",
			htmlContent:     `<html><body>No content</body></html>`,
			wantHtmlVersion: "HTML5",
			wantTitle:       "",
			wantLogin:       false,
			wantH1Count:     0,
			wantH2Count:     0,
			wantInternal:    0,
			wantExternal:    0,
		},
		{
			name:            "Test Html",
			htmlContent:     `<html lang="en"><body>No content</body></html>`,
			wantHtmlVersion: "HTML5",
		},
	}

	service := NewAnalyzerService(httpClient)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := startTestServer(tc.htmlContent)
			defer ts.Close()

			result, err := service.AnalyzePage(ts.URL)
			if err != nil {
				t.Fatalf("AnalyzePage failed: %v", err)
			}

			assertAnalyzerResult(t, result, tc)
		})
	}
}

func assertAnalyzerResult(t *testing.T, got *model.AnalyzerResult, tc analyzeTestCase) {
	t.Helper() // marking it as a test helper

	if !strings.Contains(got.PageTitle, tc.wantTitle) {
		t.Errorf("expected title to contain %q, got %q", tc.wantTitle, got.PageTitle)
	}
	if got.LoginFormDetected != tc.wantLogin {
		t.Errorf("expected loginFormDetected to be %v, got %v", tc.wantLogin, got.LoginFormDetected)
	}
	if got.Headings.H1 != tc.wantH1Count {
		t.Errorf("expected %d H1 tags, got %d", tc.wantH1Count, got.Headings.H1)
	}
	if got.Headings.H2 != tc.wantH2Count {
		t.Errorf("expected %d H2 tags, got %d", tc.wantH2Count, got.Headings.H2)
	}
	if got.InternalLinks != tc.wantInternal {
		t.Errorf("expected %d internal links, got %d", tc.wantInternal, got.InternalLinks)
	}
	if got.ExternalLinks != tc.wantExternal {
		t.Errorf("expected %d external links, got %d", tc.wantExternal, got.ExternalLinks)
	}
}

func TestAnalyzePage_InaccessibleLinks(t *testing.T) {
	simServer := simulateSuccessAndFailServer()
	defer simServer.Close()

	html := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
			<body>
				<a href="%s/ok">Accessible</a>
				<a href="%s/fail">Inaccessible</a>
			</body>
		</html>
	`, simServer.URL, simServer.URL)

	ts := startTestServer(html)
	defer ts.Close()

	service := NewAnalyzerService(httpClient)

	result, err := service.AnalyzePage(ts.URL)
	if err != nil {
		t.Fatalf("AnalyzePage failed: %v", err)
	}

	if result.ExternalLinks != 2 {
		t.Errorf("expected 2 external links, got %d", result.ExternalLinks)
	}
	if result.InaccessibleExternalLinks != 1 {
		t.Errorf("expected 1 inaccessible link, got %d", result.InaccessibleInternalLinks)
	}
}

func TestAnalyzePage_ErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantErrMsg string
	}{
		{
			name: "HTTP error status",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Forbidden", http.StatusForbidden)
			},
			wantErrMsg: "HTTP error 403",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(tc.handler)
			defer ts.Close()

			service := NewAnalyzerService(httpClient)

			_, err := service.AnalyzePage(ts.URL)
			if err == nil || !strings.Contains(err.Error(), tc.wantErrMsg) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErrMsg, err)
			}
		})
	}
}
