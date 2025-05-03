package urlanalyzer

import (
	"context"
	"github.com/sendurangr/url-analyzer-api/internal/constants"
	"github.com/sendurangr/url-analyzer-api/internal/model"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"
)

func (a *analyzerServiceImpl) checkLinksConcurrently(links []string, baseURL *url.URL, result *model.AnalyzerResult) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), constants.TimeoutSeconds*time.Second)
	defer cancel()

	type linkResult struct {
		isInternal   bool
		isAccessible bool
	}

	resultsChan := make(chan linkResult, len(links))

	// Limit the number of concurrent requests to avoid overwhelming the server
	sem := make(chan struct{}, constants.LinkCheckerConcurrentLimit)

	for _, link := range links {
		wg.Add(1)
		go func(link string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			isInternal, isAccessible := a.checkSingleLink(ctx, link, baseURL)
			resultsChan <- linkResult{isInternal, isAccessible}
		}(link)
	}

	wg.Wait()
	close(resultsChan)

	for r := range resultsChan {
		if r.isInternal {
			result.InternalLinks++
			if !r.isAccessible {
				result.InaccessibleInternalLinks++
			}
		} else {
			result.ExternalLinks++
			if !r.isAccessible {
				result.InaccessibleExternalLinks++
			}
		}
	}
}

func (a *analyzerServiceImpl) checkSingleLink(ctx context.Context, link string, baseURL *url.URL) (isInternal bool, isAccessible bool) {
	linkURL, err := url.Parse(link)
	if err != nil {
		return false, false
	}

	isInternal = linkURL.Host == "" || linkURL.Host == baseURL.Host

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, link, nil)

	if err != nil {
		return isInternal, false
	}

	resp, err := a.client.Do(req)

	if err != nil || resp.StatusCode >= 400 {
		return isInternal, false
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Error("failed to close response body", "error", err)
		}
	}(resp.Body)

	return isInternal, true
}
