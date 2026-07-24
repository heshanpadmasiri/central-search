package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/heshanpadmasiri/central-search/internal/catalog"
)

func TestSearchTextOutputIsSorted(t *testing.T) {
	var gotQuery string
	service := fakeService{searchFunc: func(_ context.Context, query string) ([]catalog.PackageSummary, error) {
		gotQuery = query
		return []catalog.PackageSummary{
			{Organization: "Wso2", Package: "GraphQL", Summary: "GraphQL tools"},
			{Organization: "ballerina", Package: "websocket", Summary: "WebSocket APIs"},
			{Organization: "ballerina", Package: "HTTP", Summary: "HTTP APIs"},
		}, nil
	}}
	var out bytes.Buffer
	root := NewRootCommand(service, IOStreams{Out: &out, ErrOut: &bytes.Buffer{}})
	root.SetArgs([]string{"search", "  HtTp  "})

	if err := root.ExecuteContext(t.Context()); err != nil {
		t.Fatalf("ExecuteContext() error = %v", err)
	}
	if gotQuery != "HtTp" {
		t.Fatalf("Search query = %q, want %q", gotQuery, "HtTp")
	}
	output := out.String()
	ordered := []string{"ballerina/HTTP", "ballerina/websocket", "Wso2/GraphQL"}
	previous := -1
	for _, value := range ordered {
		position := strings.Index(output, value)
		if position < 0 {
			t.Fatalf("output %q does not contain %q", output, value)
		}
		if position <= previous {
			t.Fatalf("output is not sorted: %q", output)
		}
		previous = position
	}
}

func TestSearchJSONOutput(t *testing.T) {
	service := fakeService{searchFunc: func(context.Context, string) ([]catalog.PackageSummary, error) {
		return []catalog.PackageSummary{{Organization: "ballerina", Package: "http", Summary: "HTTP APIs"}}, nil
	}}
	var out bytes.Buffer
	root := NewRootCommand(service, IOStreams{Out: &out, ErrOut: &bytes.Buffer{}})
	root.SetArgs([]string{"search", "http", "--json"})

	if err := root.ExecuteContext(t.Context()); err != nil {
		t.Fatalf("ExecuteContext() error = %v", err)
	}
	var got []jsonSearchResult
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON output %q: %v", out.String(), err)
	}
	want := jsonSearchResult{Organization: "ballerina", Package: "http", Summary: "HTTP APIs"}
	if len(got) != 1 || got[0] != want {
		t.Fatalf("JSON output = %#v, want %#v", got, []jsonSearchResult{want})
	}
}

func TestSearchNoMatches(t *testing.T) {
	service := fakeService{searchFunc: func(context.Context, string) ([]catalog.PackageSummary, error) {
		return nil, nil
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
			root := NewRootCommand(service, IOStreams{Out: &out, ErrOut: &bytes.Buffer{}})
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

func TestSearchRejectsEmptyQuery(t *testing.T) {
	root := NewRootCommand(fakeService{}, IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}})
	root.SetArgs([]string{"search", "  "})
	if err := root.ExecuteContext(t.Context()); err == nil || !strings.Contains(err.Error(), "must not be empty") {
		t.Fatalf("ExecuteContext() error = %v, want empty-query error", err)
	}
}

func TestSearchWrapsBackendError(t *testing.T) {
	backendErr := errors.New("network failed")
	service := fakeService{searchFunc: func(context.Context, string) ([]catalog.PackageSummary, error) {
		return nil, backendErr
	}}
	root := NewRootCommand(service, IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}})
	root.SetArgs([]string{"search", "http"})
	if err := root.ExecuteContext(t.Context()); !errors.Is(err, backendErr) {
		t.Fatalf("ExecuteContext() error = %v, want wrapped backend error", err)
	}
}
