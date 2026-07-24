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

func TestManWithoutJSONDoesNotCallService(t *testing.T) {
	root := NewRootCommand(fakeSearchService{}, fakeDocumentationService{}, IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}})
	root.SetArgs([]string{"man", "http"})
	if err := root.ExecuteContext(t.Context()); !errors.Is(err, ErrManTextRenderingUnavailable) {
		t.Fatalf("error=%v", err)
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
