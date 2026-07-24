package central

// SearchPackagesResponse is the response returned by Central's package search
// endpoint.
type SearchPackagesResponse struct {
	Packages []Package `json:"packages"`
	Count    int       `json:"count"`
	Offset   int       `json:"offset"`
	Limit    int       `json:"limit"`
}

// Package contains the package fields used by the application.
type Package struct {
	Organization string `json:"organization"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	Summary      string `json:"summary"`
}
