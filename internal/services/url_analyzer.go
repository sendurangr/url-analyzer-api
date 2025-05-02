package services

import (
	"fmt"
	"github.com/sendurangr/url-analyzer-api/internal/constants"
	"github.com/sendurangr/url-analyzer-api/internal/model"
	"golang.org/x/net/html"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type AnalyzerService interface {
	AnalyzePage(url string) (*model.AnalyzerResult, error)
}

type analyzerServiceImpl struct{}

func NewAnalyzerService() AnalyzerService {
	return &analyzerServiceImpl{}
}

var httpClient = &http.Client{
	Timeout: constants.TimeoutSeconds * time.Second,
}

func (a *analyzerServiceImpl) AnalyzePage(rawURL string) (*model.AnalyzerResult, error) {
	start := time.Now()

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	setHeaders(req)

	resp, err := httpClient.Do(req)
	if err != nil {
		slog.Error("HTTP request failed", "url", rawURL, "error", err)
		return nil, fmt.Errorf("performing request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode >= 400 {
		slog.Warn("non-OK HTTP response", "status", resp.StatusCode, "url", rawURL)
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parsing URL: %w", err)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		slog.Error("HTML parse failed", "url", rawURL, "error", err)
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	result := &model.AnalyzerResult{}
	iterateThroughDOM(doc, result, parsedURL)
	result.TimeTakenToAnalyze = float32(time.Since(start).Seconds())
	result.Url = rawURL

	return result, nil
}

func randomUserAgent() string {
	agents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
		"Mozilla/5.0 (X11; Linux x86_64)",
	}
	return agents[rand.Intn(len(agents))]
}

func setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", randomUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
}

func iterateThroughDOM(n *html.Node, result *model.AnalyzerResult, baseURL *url.URL) {
	var links []string

	var collectLinks func(*html.Node)
	collectLinks = func(n *html.Node) {

		if n.Type == html.DoctypeNode {
			extractHtmlVersionFromDoctypeNode(n, result)
		}

		if n.Type == html.ElementNode {
			switch n.Data {
			case "html":
				// when html.DoctypeNode is not present
				extractHtmlVersionFromElementNode(n, result)
			case "title":
				extractTitleFromElementNode(n, result)
			case "h1", "h2", "h3", "h4", "h5", "h6":
				countHtmlTags(n, result)
			case "a":
				extractLinksFromElementNode(n, baseURL, &links)
			case "form":
				detectLoginFormFromElementNode(n, result)
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			collectLinks(c)
		}
	}

	collectLinks(n)

	checkLinksConcurrently(links, baseURL, result)
}
