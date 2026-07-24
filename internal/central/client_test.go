package central

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchPackages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", request.Method)
		}
		if request.URL.Path != "/2.0/registry/packages/" {
			t.Errorf("path = %q", request.URL.Path)
		}
		if got := request.URL.Query().Get("q"); got != "http client" {
			t.Errorf("q = %q, want %q", got, "http client")
		}
		if got := request.URL.Query().Get("limit"); got != "2" {
			t.Errorf("limit = %q, want 2", got)
		}
		if got := request.URL.Query().Get("offset"); got != "3" {
			t.Errorf("offset = %q, want 3", got)
		}
		if got := request.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q", got)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"packages":[{"organization":"ballerina","name":"http","version":"2.16.4","summary":"HTTP APIs","ignored":"value"}],"count":68,"offset":0,"limit":2}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL+"/2.0/registry", server.Client())
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	limit, offset := 2, 3
	result, err := client.SearchPackages(t.Context(), "http client", SearchPackagesOptions{Limit: &limit, Offset: &offset})
	if err != nil {
		t.Fatalf("SearchPackages() error = %v", err)
	}
	if result.Count != 68 || result.Offset != 0 || result.Limit != 2 || len(result.Packages) != 1 {
		t.Fatalf("SearchPackages() = %#v", result)
	}
	want := (Package{Organization: "ballerina", Name: "http", Version: "2.16.4", Summary: "HTTP APIs"})
	if result.Packages[0] != want {
		t.Fatalf("package = %#v, want %#v", result.Packages[0], want)
	}
}

func TestSearchPackagesOmitsLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if _, exists := request.URL.Query()["limit"]; exists {
			t.Errorf("query unexpectedly contains limit: %q", request.URL.RawQuery)
		}
		if _, exists := request.URL.Query()["offset"]; exists {
			t.Errorf("query unexpectedly contains offset: %q", request.URL.RawQuery)
		}
		_, _ = response.Write([]byte(`{"packages":[],"count":0,"offset":0,"limit":15}`))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	result, err := client.SearchPackages(t.Context(), "missing", SearchPackagesOptions{})
	if err != nil {
		t.Fatalf("SearchPackages() error = %v", err)
	}
	if len(result.Packages) != 0 {
		t.Fatalf("packages = %#v, want empty", result.Packages)
	}
}

func TestSearchPackagesReportsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		http.Error(response, `{"message":"bad query"}`, http.StatusBadRequest)
	}))
	defer server.Close()
	client, _ := NewClient(server.URL, server.Client())
	_, err := client.SearchPackages(t.Context(), "http", SearchPackagesOptions{})
	if err == nil || !strings.Contains(err.Error(), "400 Bad Request") || !strings.Contains(err.Error(), "bad query") {
		t.Fatalf("SearchPackages() error = %v", err)
	}
}

func TestSearchPackagesReportsInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte(`not-json`))
	}))
	defer server.Close()
	client, _ := NewClient(server.URL, server.Client())
	_, err := client.SearchPackages(t.Context(), "http", SearchPackagesOptions{})
	if err == nil || !strings.Contains(err.Error(), "decode package search response") {
		t.Fatalf("SearchPackages() error = %v", err)
	}
}

func TestSearchPackagesHonorsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	client, err := NewClient("https://example.com", nil)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	_, err = client.SearchPackages(ctx, "http", SearchPackagesOptions{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("SearchPackages() error = %v, want context.Canceled", err)
	}
}

func TestPackageVersionsPreservesOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/registry/packages/ballerina/http" {
			t.Errorf("path = %q", request.URL.Path)
		}
		if request.Header.Get("Accept") != "application/json" {
			t.Errorf("Accept = %q", request.Header.Get("Accept"))
		}
		_, _ = response.Write([]byte(`["2.0.0","1.0.0"]`))
	}))
	defer server.Close()
	client, _ := NewClient(server.URL+"/registry", server.Client())
	versions, err := client.PackageVersions(t.Context(), "ballerina", "http")
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 2 || versions[0] != "2.0.0" || versions[1] != "1.0.0" {
		t.Fatalf("versions = %#v", versions)
	}
}

func TestPackageVersionsErrors(t *testing.T) {
	for name, handler := range map[string]http.HandlerFunc{
		"status": func(response http.ResponseWriter, _ *http.Request) {
			http.Error(response, "failed", http.StatusBadGateway)
		},
		"json": func(response http.ResponseWriter, _ *http.Request) { _, _ = response.Write([]byte("not-json")) },
	} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(handler)
			defer server.Close()
			client, _ := NewClient(server.URL, server.Client())
			if _, err := client.PackageVersions(t.Context(), "a", "b"); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestNewClientRejectsInvalidBaseURL(t *testing.T) {
	for _, baseURL := range []string{"central.example.com", "https://example.com?query=value", "://bad"} {
		t.Run(baseURL, func(t *testing.T) {
			if _, err := NewClient(baseURL, nil); err == nil {
				t.Fatalf("NewClient(%q) succeeded, want error", baseURL)
			}
		})
	}
}
