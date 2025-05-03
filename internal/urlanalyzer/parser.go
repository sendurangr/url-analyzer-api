package urlanalyzer

import (
	"github.com/sendurangr/url-analyzer-api/internal/constants"
	"github.com/sendurangr/url-analyzer-api/internal/model"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"net/url"
	"strings"
)

func extractHtmlVersionFromDoctypeNode(n *html.Node, result *model.AnalyzerResult) {

	if n.DataAtom == atom.Html || strings.EqualFold(n.Data, "html") {
		result.HTMLVersion = constants.HTML5Version
	} else {
		result.HTMLVersion = constants.LegacyHTMLVersion
	}
}

func extractHtmlVersionFromElementNode(n *html.Node, result *model.AnalyzerResult) {

	// already set by doctype node
	if result.HTMLVersion != "" {
		return
	}

	for _, attr := range n.Attr {
		if attr.Key == "lang" {
			result.HTMLVersion = constants.HTML5Version
			return
		}
	}

	result.HTMLVersion = constants.LegacyHTMLVersion
}

func extractTitleFromElementNode(n *html.Node, result *model.AnalyzerResult) {
	if n.FirstChild != nil {
		result.PageTitle = n.FirstChild.Data
	}
}

func extractLinksFromElementNode(n *html.Node, baseURL *url.URL, links *[]string) {
	for _, attr := range n.Attr {
		if attr.Key != "href" || attr.Val == "" || strings.HasPrefix(attr.Val, "#") {
			continue
		}

		if linkURL, err := url.Parse(attr.Val); err == nil {
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

		// early exit when both fields are found - no need to traverse further
		if hasPassword && hasUserField {
			return
		}

		if n.Type == html.ElementNode && n.DataAtom == atom.Input {
			var inputType string
			for _, attr := range n.Attr {
				if attr.Key == "type" {
					inputType = strings.ToLower(attr.Val)
					break
				}
			}

			switch inputType {
			case "password":
				hasPassword = true
			case "text", "email":
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
