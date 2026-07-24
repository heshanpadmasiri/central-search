// Package central provides an HTTP client for the Ballerina Central API.
package central

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the public Ballerina Central registry API URL.
	DefaultBaseURL = "https://api.central.ballerina.io/2.0/registry"
	maxErrorBody   = 8 << 10
)

// SearchPackagesOptions controls a package search request.
type SearchPackagesOptions struct {
	// Limit is omitted from the request when nil, allowing Central to apply its
	// default.
	Limit *int
}

// Client calls the Ballerina Central API.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

// NewClient constructs a Central API client. When httpClient is nil, a client
// with a 30-second timeout is used.
func NewClient(baseURL string, httpClient *http.Client) (*Client, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse Central base URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("parse Central base URL: URL must include scheme and host")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("parse Central base URL: query and fragment are not allowed")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/"
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{baseURL: parsed, httpClient: httpClient}, nil
}

// SearchPackages searches Central for packages matching query.
func (c *Client) SearchPackages(ctx context.Context, query string, options SearchPackagesOptions) (SearchPackagesResponse, error) {
	endpoint := *c.baseURL
	endpoint.Path += "packages/"
	parameters := endpoint.Query()
	parameters.Set("q", query)
	if options.Limit != nil {
		parameters.Set("limit", fmt.Sprintf("%d", *options.Limit))
	}
	endpoint.RawQuery = parameters.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return SearchPackagesResponse{}, fmt.Errorf("create package search request: %w", err)
	}
	request.Header.Set("Accept", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return SearchPackagesResponse{}, fmt.Errorf("send package search request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, readErr := io.ReadAll(io.LimitReader(response.Body, maxErrorBody))
		if readErr != nil {
			return SearchPackagesResponse{}, fmt.Errorf("Central package search returned %s (read error response: %v)", response.Status, readErr)
		}
		message := strings.TrimSpace(string(body))
		if message == "" {
			return SearchPackagesResponse{}, fmt.Errorf("Central package search returned %s", response.Status)
		}
		return SearchPackagesResponse{}, fmt.Errorf("Central package search returned %s: %s", response.Status, message)
	}

	var result SearchPackagesResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return SearchPackagesResponse{}, fmt.Errorf("decode package search response: %w", err)
	}
	return result, nil
}
