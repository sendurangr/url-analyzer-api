package model

type AnalyzerResult struct {
	HTMLVersion       string   `json:"htmlVersion"`
	PageTitle         string   `json:"pageTitle"`
	Headings          Headings `json:"headings"`
	InternalLinks     int      `json:"internalLinks"`
	ExternalLinks     int      `json:"externalLinks"`
	InaccessibleLinks int      `json:"inaccessibleLinks"`
	LoginFormDetected bool     `json:"loginFormDetected"`
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
