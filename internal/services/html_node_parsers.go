package services

import (
	"github.com/sendurangr/url-analyzer-api/internal/model"
	"golang.org/x/net/html"
	"net/url"
	"strings"
)

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
