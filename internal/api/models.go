package api

type BasicSearchTerms struct {
	Include []string `json:"include"`
	Exclude []string `json:"exclude,omitempty"`
}

type Request struct {
	APIKey           string           `json:"apiKey"`
	SearchType       string           `json:"searchType,omitempty"` // "current" or "historic"
	Mode             string           `json:"mode,omitempty"`       // "preview" or "purchase"
	Punycode         bool             `json:"punycode"`
	BasicSearchTerms BasicSearchTerms `json:"basicSearchTerms"`
}

type Response struct {
	NextPageSearchAfter any      `json:"nextPageSearchAfter"`
	DomainsCount        int      `json:"domainsCount"`
	DomainsList         []string `json:"domainsList"`
}
