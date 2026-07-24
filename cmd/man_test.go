package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/heshanpadmasiri/central-search/internal/catalog"
)

func TestManWritesOpaqueDocumentation(t *testing.T) {
	var gotSelector catalog.PackageSelector
	service := fakeDocumentationService{documentationFunc: func(_ context.Context, selector catalog.PackageSelector) (catalog.PackageDocumentation, error) {
		gotSelector = selector
		return catalog.PackageDocumentation{Content: []byte("# HTTP\n\nPackage documentation.\n")}, nil
	}}
	var out bytes.Buffer
	root := NewRootCommand(fakeSearchService{}, service, IOStreams{Out: &out, ErrOut: &bytes.Buffer{}})
	root.SetArgs([]string{"man", "ballerina/http"})

	if err := root.ExecuteContext(t.Context()); err != nil {
		t.Fatalf("ExecuteContext() error = %v", err)
	}
	wantSelector := catalog.PackageSelector{Organization: "ballerina", Package: "http"}
	if gotSelector != wantSelector {
		t.Fatalf("Documentation selector = %#v, want %#v", gotSelector, wantSelector)
	}
	if got := out.String(); got != "# HTTP\n\nPackage documentation.\n" {
		t.Fatalf("output = %q", got)
	}
}

func TestManAcceptsUnqualifiedPackage(t *testing.T) {
	service := fakeDocumentationService{documentationFunc: func(_ context.Context, selector catalog.PackageSelector) (catalog.PackageDocumentation, error) {
		if selector.Organization != "" || selector.Package != "http" {
			t.Fatalf("selector = %#v", selector)
		}
		return catalog.PackageDocumentation{}, nil
	}}
	root := NewRootCommand(fakeSearchService{}, service, IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}})
	root.SetArgs([]string{"man", "http"})
	if err := root.ExecuteContext(t.Context()); err != nil {
		t.Fatalf("ExecuteContext() error = %v", err)
	}
}

func TestManRejectsInvalidPackage(t *testing.T) {
	root := NewRootCommand(fakeSearchService{}, fakeDocumentationService{}, IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}})
	root.SetArgs([]string{"man", "ballerina/http/client"})
	if err := root.ExecuteContext(t.Context()); err == nil {
		t.Fatal("ExecuteContext() succeeded, want invalid-package error")
	}
}

func TestManWrapsBackendError(t *testing.T) {
	backendErr := errors.New("network failed")
	service := fakeDocumentationService{documentationFunc: func(context.Context, catalog.PackageSelector) (catalog.PackageDocumentation, error) {
		return catalog.PackageDocumentation{}, backendErr
	}}
	root := NewRootCommand(fakeSearchService{}, service, IOStreams{Out: &bytes.Buffer{}, ErrOut: &bytes.Buffer{}})
	root.SetArgs([]string{"man", "http"})
	if err := root.ExecuteContext(t.Context()); !errors.Is(err, backendErr) {
		t.Fatalf("ExecuteContext() error = %v, want wrapped backend error", err)
	}
}
