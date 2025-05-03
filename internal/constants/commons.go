package constants

import "time"

const (
	LinkCheckerConcurrentLimit = 64
	HTML5Version               = "HTML5"
	LegacyHTMLVersion          = "Older HTML or XHTML"
)

const (
	HttpClientTimeout = 15 * time.Second
	ContextTimeout    = 20 * time.Second
)
