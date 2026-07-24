package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/heshanpadmasiri/central-search/internal/search"
)

func TestSearchTextOutputPreservesRelevanceOrder(t *testing.T) {
	var gotQuery string
	var gotOptions search.Options
	service := fakeSearchService{searchFunc: func(_ context.Context, query string, options search.Options) (search.Response, error) {
		gotQuery = query
		gotOptions = options
		return search.Response{Packages: []search.Package{
			{Organization: "Wso2", Package: "GraphQL", Version: "2.0.0", Summary: "GraphQL tools"},
			{Organization: "ballerina", Package: "HTTP", Version: "2.16.4", Summary: "HTTP APIs"},
		}}, nil
	}}
	var out bytes.Buffer
	root := NewRootCommand(service, fakeDocumentationService{}, IOStreams{Out: &out, ErrOut: &bytes.Buffer{}})
	root.SetArgs([]string{"search", "  HtTp  ", "--limit", "7"})

	if err := root.ExecuteContext(t.Context()); err != nil {
		t.Fatalf("ExecuteContext() error = %v", err)
	}
	if gotQuery != "HtTp" {
		t.Fatalf("Search query = %q, want %q", gotQuery, "HtTp")
	}
	if gotOptions.Limit == nil || *gotOptions.Limit != 7 {
		t.Fatalf("Search limit = %v, want 7", gotOptions.Limit)
	}
	output := out.String()
	first := strings.Index(output, "Wso2/GraphQL")
	second := strings.Index(output, "ballerina/HTTP")
	if first < 0 || second <= first {
		t.Fatalf("output does not preserve service order: %q", output)
	}
	if !strings.Contains(output, "2.16.4") {
		t.Fatalf("output does not contain version: %q", output)
	}
}

func TestSearchOmitsUnspecifiedLimit(t *testing.T) {
	service := fakeSearchService{searchFunc: func(_ context.Context, _ string, options search.Options) (search.Response, error) {
		if options.Limit != nil {
			t.Fatalf("Search limit = %d, want nil", *options.Limit)
		}
		return search.Response{Packages: []search.Package{{Package: "http"}}}, nil
	}}
	root := NewRootCommand(service, fakeDocumentationService{}, IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}})
	root.SetArgs([]string{"search", "http"})
	if err := root.ExecuteContext(t.Context()); err != nil {
		t.Fatalf("ExecuteContext() error = %v", err)
	}
}

func TestSearchJSONOutput(t *testing.T) {
	service := fakeSearchService{searchFunc: func(context.Context, string, search.Options) (search.Response, error) {
		return search.Response{Packages: []search.Package{{Organization: "ballerina", Package: "http", Version: "2.16.4", Summary: "HTTP APIs"}}}, nil
	}}
	var out bytes.Buffer
	root := NewRootCommand(service, fakeDocumentationService{}, IOStreams{Out: &out, ErrOut: &bytes.Buffer{}})
	root.SetArgs([]string{"search", "http", "--json"})

	if err := root.ExecuteContext(t.Context()); err != nil {
		t.Fatalf("ExecuteContext() error = %v", err)
	}
	var got []jsonSearchResult
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON output %q: %v", out.String(), err)
	}
	want := jsonSearchResult{Organization: "ballerina", Package: "http", Version: "2.16.4", Summary: "HTTP APIs"}
	if len(got) != 1 || got[0] != want {
		t.Fatalf("JSON output = %#v, want %#v", got, []jsonSearchResult{want})
	}
}

func TestSearchNoMatches(t *testing.T) {
	service := fakeSearchService{searchFunc: func(context.Context, string, search.Options) (search.Response, error) {
		return search.Response{}, nil
	}}
	for _, test := range []struct {
		name    string
		args    []string
		wantOut string
	}{
		{name: "text", args: []string{"search", "missing"}},
		{name: "json", args: []string{"search", "missing", "--json"}, wantOut: "[]\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var out bytes.Buffer
			root := NewRootCommand(service, fakeDocumentationService{}, IOStreams{Out: &out, ErrOut: &bytes.Buffer{}})
			root.SetArgs(test.args)
			err := root.ExecuteContext(t.Context())
			if !errors.Is(err, ErrNoMatches) {
				t.Fatalf("ExecuteContext() error = %v, want ErrNoMatches", err)
			}
			if out.String() != test.wantOut {
				t.Fatalf("output = %q, want %q", out.String(), test.wantOut)
			}
		})
	}
}

func TestSearchRejectsInvalidInput(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
		want error
	}{
		{name: "empty query", args: []string{"search", "  "}, want: search.ErrEmptyQuery},
		{name: "zero limit", args: []string{"search", "http", "--limit", "0"}, want: search.ErrInvalidLimit},
		{name: "negative limit", args: []string{"search", "http", "--limit", "-1"}, want: search.ErrInvalidLimit},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := NewRootCommand(fakeSearchService{}, fakeDocumentationService{}, IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}})
			root.SetArgs(test.args)
			if err := root.ExecuteContext(t.Context()); !errors.Is(err, test.want) {
				t.Fatalf("ExecuteContext() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestSearchWrapsBackendError(t *testing.T) {
	backendErr := errors.New("network failed")
	service := fakeSearchService{searchFunc: func(context.Context, string, search.Options) (search.Response, error) {
		return search.Response{}, backendErr
	}}
	root := NewRootCommand(service, fakeDocumentationService{}, IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}})
	root.SetArgs([]string{"search", "http"})
	if err := root.ExecuteContext(t.Context()); !errors.Is(err, backendErr) {
		t.Fatalf("ExecuteContext() error = %v, want wrapped backend error", err)
	}
}
