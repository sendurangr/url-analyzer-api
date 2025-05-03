package urlanalyzer

import (
	"fmt"
	"github.com/sendurangr/url-analyzer-api/internal/model"
	"github.com/sendurangr/url-analyzer-api/internal/utils"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type AnalyzerService interface {
	AnalyzePage(url string) (*model.AnalyzerResult, error)
}

type analyzerServiceImpl struct {
	client *http.Client
}

func NewAnalyzerService(client *http.Client) AnalyzerService {
	return &analyzerServiceImpl{client: client}
}

func (a *analyzerServiceImpl) AnalyzePage(rawURL string) (*model.AnalyzerResult, error) {
	start := time.Now()

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	utils.SetHeaders(req)

	resp, err := a.client.Do(req)

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
	a.iterateThroughDOM(doc, result, parsedURL)
	result.TimeTakenToAnalyze = float32(time.Since(start).Seconds())
	result.Url = rawURL

	return result, nil
}

func (a *analyzerServiceImpl) iterateThroughDOM(n *html.Node, result *model.AnalyzerResult, baseURL *url.URL) {
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
