package model

type AnalyzerResult struct {
	HTMLVersion       string   `json:"html_version"`
	PageTitle         string   `json:"page_title"`
	Headings          Headings `json:"headings"`
	InternalLinks     int      `json:"internal_links"`
	ExternalLinks     int      `json:"external_links"`
	InaccessibleLinks int      `json:"inaccessible_links"`
	LoginFormDetected bool     `json:"login_form_detected"`
}

type Headings struct {
	H1 int `json:"h1"`
	H2 int `json:"h2"`
	H3 int `json:"h3"`
	H4 int `json:"h4"`
	H5 int `json:"h5"`
	H6 int `json:"h6"`
	H7 int `json:"h7"`
	H8 int `json:"h8"`
}
