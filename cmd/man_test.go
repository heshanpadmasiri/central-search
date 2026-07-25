package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/heshanpadmasiri/central-search/internal/catalog"
)

func TestManWritesJSONAndWarnings(t *testing.T) {
	var gotQuery string
	service := fakeDocumentationService{documentationFunc: func(_ context.Context, query string) (catalog.Package, error) {
		gotQuery = query
		return catalog.Package{SchemaVersion: 1, Complete: false, Warnings: []catalog.Warning{{Module: "http.extra", Message: "unavailable"}}, Modules: []catalog.Module{}}, nil
	}}
	var out, errOut bytes.Buffer
	root := NewRootCommand(fakeSearchService{}, service, IOStreams{Out: &out, ErrOut: &errOut})
	root.SetArgs([]string{"man", "  http client  ", "--json"})
	if err := root.ExecuteContext(t.Context()); err != nil {
		t.Fatal(err)
	}
	if gotQuery != "http client" {
		t.Fatalf("query=%q", gotQuery)
	}
	var result catalog.Package
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON %q: %v", out.String(), err)
	}
	if errOut.String() != "Warning: http.extra: unavailable\n" {
		t.Fatalf("stderr=%q", errOut.String())
	}
}

func TestManWritesMarkdownWithoutJSON(t *testing.T) {
	service := fakeDocumentationService{documentationFunc: func(_ context.Context, query string) (catalog.Package, error) {
		if query != "http" {
			t.Fatalf("query=%q", query)
		}
		return catalog.Package{
			Organization: "ballerina", Name: "http", Version: "2.16.4", Complete: true,
			Modules: []catalog.Module{{Name: "http", IsDefault: true, Functions: []catalog.Function{{Symbol: catalog.Symbol{Name: "parseHeader", Signature: "function parseHeader(string value)"}}}}},
		}, nil
	}}
	var out bytes.Buffer
	root := NewRootCommand(fakeSearchService{}, service, IOStreams{Out: &out, ErrOut: &bytes.Buffer{}})
	root.SetArgs([]string{"man", "http"})
	if err := root.ExecuteContext(t.Context()); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"# ballerina/http", "**Version:** 2.16.4", "## Module: http", "### Functions", "#### `parseHeader`", "```ballerina"} {
		if !bytes.Contains(out.Bytes(), []byte(expected)) {
			t.Fatalf("output does not contain %q:\n%s", expected, out.String())
		}
	}
	if bytes.Contains(out.Bytes(), []byte("\x1b[")) {
		t.Fatalf("piped output contains ANSI escapes: %q", out.String())
	}
}

func TestManWrapsBackendError(t *testing.T) {
	backendErr := errors.New("network failed")
	service := fakeDocumentationService{documentationFunc: func(context.Context, string) (catalog.Package, error) { return catalog.Package{}, backendErr }}
	root := NewRootCommand(fakeSearchService{}, service, IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}})
	root.SetArgs([]string{"man", "http", "--json"})
	if err := root.ExecuteContext(t.Context()); !errors.Is(err, backendErr) {
		t.Fatalf("error=%v", err)
	}
}
