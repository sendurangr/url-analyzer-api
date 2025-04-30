package services

import (
	"context"
	"errors"
	"github.com/sendurangr/url-analyzer-api/internal/constants"
	"github.com/sendurangr/url-analyzer-api/internal/model"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var httpClient = &http.Client{
	Timeout: constants.TimeoutSeconds * time.Second,
}

func AnalyzePage(rawURL string) (*model.AnalyzerResult, error) {
	resp, err := httpClient.Get(rawURL)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(resp.Body)

	if resp.StatusCode >= 400 {
		return nil, errors.New("Failed to fetch page: HTTP " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	result := &model.AnalyzerResult{}

	iterateThroughDOM(doc, result, parsedURL)

	return result, nil
}

func iterateThroughDOM(n *html.Node, result *model.AnalyzerResult, baseURL *url.URL) {
	var links []string

	var collectLinks func(*html.Node)
	collectLinks = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "html":
				extractHtmlVersionFromNode(n, result)
			case "title":
				if n.FirstChild != nil {
					result.PageTitle = n.FirstChild.Data
				}
			case "h1", "h2", "h3", "h4", "h5", "h6":
				countHtmlTags(n, result)
			case "a":
				extractLinksFromNode(n, baseURL, &links)
			case "form":
				for _, attr := range n.Attr {
					if attr.Key == "action" && strings.Contains(strings.ToLower(attr.Val), "login") {
						result.LoginFormDetected = true
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			collectLinks(c)
		}
	}

	collectLinks(n)

	checkLinksConcurrently(links, baseURL, result)
}

func extractHtmlVersionFromNode(n *html.Node, result *model.AnalyzerResult) {

	if result.HTMLVersion != "" {
		return
	}

	if n.Type == html.DoctypeNode {
		if strings.EqualFold(n.Data, "html") {
			result.HTMLVersion = "HTML5"
			return
		} else {
			result.HTMLVersion = "Older HTML or XHTML"
			return
		}
	}

	if n.Type == html.ElementNode && n.Data == "html" {
		for _, attr := range n.Attr {
			if attr.Key == "lang" {
				result.HTMLVersion = "HTML5"
				return
			}
		}
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

func extractLinksFromNode(n *html.Node, baseURL *url.URL, links *[]string) {
	for _, attr := range n.Attr {
		if attr.Key == "href" {
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
