package urlanalyzer

import (
	"context"
	"fmt"
	"github.com/sendurangr/url-analyzer-api/internal/constants"
	"github.com/sendurangr/url-analyzer-api/internal/model"
	"github.com/sendurangr/url-analyzer-api/internal/utils"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// AnalyzerService Interface Definition for AnalyzerService
type AnalyzerService interface {
	AnalyzePage(url string) (*model.AnalyzerResult, error)
}

// AnalyzerService implementation
type analyzer struct {
	client *http.Client
}

// NewAnalyzer DI constructor for AnalyzerService
func NewAnalyzer(client *http.Client) AnalyzerService {
	return &analyzer{client: client}
}

// AnalyzePage fetches the HTML content of the given URL and analyzes it for various attributes.
func (a *analyzer) AnalyzePage(rawURL string) (*model.AnalyzerResult, error) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), constants.ContextTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	utils.SetHeaders(req)

	resp, err := a.client.Do(req)
	if err != nil {
		slog.Error("HTTP request failed", "url", rawURL, "error", err)
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("Failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode >= 400 {
		slog.Warn("Non-OK HTTP response", "url", rawURL, "status", resp.StatusCode)
		return nil, fmt.Errorf("HTTP error %d: %s — the URL is unreachable or returned an error",
			resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		slog.Error("Failed to parse HTML", "url", rawURL, "error", err)
		return nil, fmt.Errorf("failed to parse the HTML document: %w", err)
	}

	result := &model.AnalyzerResult{}
	a.iterateThroughDOM(doc, result, parsedURL)
	result.TimeTakenToAnalyze = float32(time.Since(start).Seconds())
	result.URL = rawURL

	return result, nil
}

func (a *analyzer) iterateThroughDOM(n *html.Node, result *model.AnalyzerResult, baseURL *url.URL) {
	var links []string

	var collectLinks func(*html.Node)
	collectLinks = func(n *html.Node) {

		if n.Type == html.DoctypeNode {
			extractHtmlVersionFromDoctypeNode(n, result)
		}

		if n.Type == html.ElementNode {
			switch n.DataAtom {
			case atom.Html:
				// when html.DoctypeNode is not present in the html doc
				extractHtmlVersionFromElementNode(n, result)
			case atom.Title:
				extractTitleFromElementNode(n, result)
			case atom.A:
				extractLinksFromElementNode(n, baseURL, &links)
			case atom.Form:
				detectLoginFormFromElementNode(n, result)
			case atom.H1:
				result.Headings.H1++
			case atom.H2:
				result.Headings.H2++
			case atom.H3:
				result.Headings.H3++
			case atom.H4:
				result.Headings.H4++
			case atom.H5:
				result.Headings.H5++
			case atom.H6:
				result.Headings.H6++
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			collectLinks(c)
		}
	}

	collectLinks(n)

	a.checkLinksConcurrently(links, baseURL, result)
}
