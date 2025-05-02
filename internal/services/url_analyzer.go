package services

import (
	"context"
	"fmt"
	"github.com/sendurangr/url-analyzer-api/internal/constants"
	"github.com/sendurangr/url-analyzer-api/internal/model"
	"golang.org/x/net/html"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// AnalyzerService DI for AnalyzerService
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

func extractHtmlVersionFromDoctypeNode(n *html.Node, result *model.AnalyzerResult) {
	if result.HTMLVersion != "" {
		return
	}

	if strings.EqualFold(n.Data, "html") {
		result.HTMLVersion = "HTML5"
	} else {
		result.HTMLVersion = "Older HTML or XHTML"
	}
}

func extractHtmlVersionFromElementNode(n *html.Node, result *model.AnalyzerResult) {

	if result.HTMLVersion != "" {
		return
	}

	for _, attr := range n.Attr {
		if attr.Key == "lang" {
			result.HTMLVersion = "HTML5"
			return
		}
	}
}

func extractTitleFromElementNode(n *html.Node, result *model.AnalyzerResult) {
	if n.FirstChild != nil {
		result.PageTitle = n.FirstChild.Data
	}
}

func countHtmlTags(n *html.Node, result *model.AnalyzerResult) {
	switch n.Data {
	case "h1":
		result.Headings.H1++
	case "h2":
		result.Headings.H2++
	case "h3":
		result.Headings.H3++
	case "h4":
		result.Headings.H4++
	case "h5":
		result.Headings.H5++
	case "h6":
		result.Headings.H6++
	}
}

func extractLinksFromElementNode(n *html.Node, baseURL *url.URL, links *[]string) {
	for _, attr := range n.Attr {
		if attr.Key != "href" {
			continue
		}

		if strings.HasPrefix(attr.Val, "#") || attr.Val == "" {
			continue
		}

		linkURL, err := url.Parse(attr.Val)

		if err == nil {
			absURL := baseURL.ResolveReference(linkURL)
			*links = append(*links, absURL.String())
		}
	}
}

func detectLoginFormFromElementNode(n *html.Node, result *model.AnalyzerResult) {

	var hasPassword bool
	var hasUserField bool

	// checking only type inside form node, not other fields like <input name="username"> to support multi-language
	var checkInputs func(*html.Node)
	checkInputs = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "input" {
			var inputType string
			for _, attr := range n.Attr {

				if strings.ToLower(attr.Key) == "type" {
					inputType = strings.ToLower(attr.Val)
				}
			}
			if inputType == "password" {
				hasPassword = true
			}
			if inputType == "text" || inputType == "email" {
				hasUserField = true
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			checkInputs(c)
		}
	}

	checkInputs(n)

	if hasPassword && hasUserField {
		result.LoginFormDetected = true
	}
}

func checkLinksConcurrently(links []string, baseURL *url.URL, result *model.AnalyzerResult) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutSeconds*time.Second)
	defer cancel()

	type linkResult struct {
		isInternal   bool
		isAccessible bool
	}

	resultsChannel := make(chan linkResult, len(links))

	for _, link := range links {
		wg.Add(1)
		go func(link string) {
			defer wg.Done()
			isInternal, isAccessible := checkSingleLink(ctx, link, baseURL)
			resultsChannel <- linkResult{isInternal, isAccessible}
		}(link)
	}

	wg.Wait()
	close(resultsChannel)

	for r := range resultsChannel {
		if r.isInternal {
			result.InternalLinks++
		} else {
			result.ExternalLinks++
		}
		if !r.isAccessible {
			result.InaccessibleLinks++
		}
	}
}

func checkSingleLink(ctx context.Context, link string, baseURL *url.URL) (isInternal bool, isAccessible bool) {
	linkURL, err := url.Parse(link)
	if err != nil {
		return false, false
	}

	isInternal = linkURL.Host == "" || linkURL.Host == baseURL.Host

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, link, nil)

	if err != nil {
		return isInternal, false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		return isInternal, false
	}

	return isInternal, true
}
