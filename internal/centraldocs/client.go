// Package centraldocs scrapes structured documentation embedded in Central pages.
package centraldocs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/heshanpadmasiri/central-search/internal/catalog"
)

const (
	DefaultBaseURL = "https://central.ballerina.io"
	maxHTMLSize    = 32 << 20
	maxErrorBody   = 8 << 10
)

// Client scrapes structured API documentation embedded in Central pages.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

// NewClient constructs a Central documentation scraper.
func NewClient(baseURL string, httpClient *http.Client) (*Client, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse Central documentation base URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("parse Central documentation base URL: URL must include scheme and host")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("parse Central documentation base URL: query and fragment are not allowed")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/"
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{baseURL: parsed, httpClient: httpClient}, nil
}

// ScrapeModule fetches and normalizes one module page.
func (c *Client) ScrapeModule(ctx context.Context, organization, module, version string) (catalog.ScrapedModule, error) {
	endpoint := *c.baseURL
	endpoint.Path += url.PathEscape(organization) + "/" + url.PathEscape(module) + "/" + url.PathEscape(version)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return catalog.ScrapedModule{}, fmt.Errorf("create Central module page request: %w", err)
	}
	request.Header.Set("Accept", "text/html")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return catalog.ScrapedModule{}, fmt.Errorf("fetch Central module page: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, readErr := io.ReadAll(io.LimitReader(response.Body, maxErrorBody))
		if readErr != nil {
			return catalog.ScrapedModule{}, fmt.Errorf("Central module page returned %s (read error response: %v)", response.Status, readErr)
		}
		message := strings.TrimSpace(string(body))
		if message == "" {
			return catalog.ScrapedModule{}, fmt.Errorf("Central module page returned %s", response.Status)
		}
		return catalog.ScrapedModule{}, fmt.Errorf("Central module page returned %s: %s", response.Status, message)
	}
	htmlData, err := io.ReadAll(io.LimitReader(response.Body, maxHTMLSize+1))
	if err != nil {
		return catalog.ScrapedModule{}, fmt.Errorf("read Central module page: %w", err)
	}
	if len(htmlData) > maxHTMLSize {
		return catalog.ScrapedModule{}, fmt.Errorf("Central module page exceeds %d bytes", maxHTMLSize)
	}
	data, err := extractNextData(bytes.NewReader(htmlData))
	if err != nil {
		return catalog.ScrapedModule{}, fmt.Errorf("extract Central module data: %w", err)
	}
	var page nextData
	if err := json.Unmarshal(data, &page); err != nil {
		return catalog.ScrapedModule{}, fmt.Errorf("decode Central module data: %w", err)
	}
	result, err := normalizePage(page.Props.PageProps.PropsData.Props)
	if err != nil {
		return catalog.ScrapedModule{}, err
	}
	return result, nil
}
