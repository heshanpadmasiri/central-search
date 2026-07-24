package search

import (
	"context"
	"errors"
	"testing"

	"github.com/heshanpadmasiri/central-search/internal/central"
)

type fakePackageClient struct {
	searchFunc func(context.Context, string, central.SearchPackagesOptions) (central.SearchPackagesResponse, error)
}

func (f fakePackageClient) SearchPackages(ctx context.Context, query string, options central.SearchPackagesOptions) (central.SearchPackagesResponse, error) {
	return f.searchFunc(ctx, query, options)
}

func TestSearchMapsResponseAndPreservesOrder(t *testing.T) {
	var gotQuery string
	var gotLimit *int
	client := fakePackageClient{searchFunc: func(_ context.Context, query string, options central.SearchPackagesOptions) (central.SearchPackagesResponse, error) {
		gotQuery = query
		gotLimit = options.Limit
		return central.SearchPackagesResponse{Packages: []central.Package{
			{Organization: "wso2", Name: "second", Version: "2.0.0", Summary: "Second"},
			{Organization: "ballerina", Name: "first", Version: "1.0.0", Summary: "First"},
		}}, nil
	}}
	limit := 5
	response, err := NewService(client).Search(t.Context(), "  http  ", Options{Limit: &limit})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if gotQuery != "http" || gotLimit == nil || *gotLimit != 5 {
		t.Fatalf("client arguments = query %q, limit %v", gotQuery, gotLimit)
	}
	want := []Package{
		{Organization: "wso2", Package: "second", Version: "2.0.0", Summary: "Second"},
		{Organization: "ballerina", Package: "first", Version: "1.0.0", Summary: "First"},
	}
	if len(response.Packages) != len(want) {
		t.Fatalf("packages = %#v, want %#v", response.Packages, want)
	}
	for i := range want {
		if response.Packages[i] != want[i] {
			t.Fatalf("package %d = %#v, want %#v", i, response.Packages[i], want[i])
		}
	}
}

func TestSearchRejectsInvalidInput(t *testing.T) {
	service := NewService(fakePackageClient{searchFunc: func(context.Context, string, central.SearchPackagesOptions) (central.SearchPackagesResponse, error) {
		panic("unexpected client call")
	}})
	zero := 0
	if _, err := service.Search(t.Context(), " ", Options{}); !errors.Is(err, ErrEmptyQuery) {
		t.Fatalf("empty Search() error = %v, want ErrEmptyQuery", err)
	}
	if _, err := service.Search(t.Context(), "http", Options{Limit: &zero}); !errors.Is(err, ErrInvalidLimit) {
		t.Fatalf("zero-limit Search() error = %v, want ErrInvalidLimit", err)
	}
}

func TestSearchWrapsClientError(t *testing.T) {
	clientErr := errors.New("network failed")
	service := NewService(fakePackageClient{searchFunc: func(context.Context, string, central.SearchPackagesOptions) (central.SearchPackagesResponse, error) {
		return central.SearchPackagesResponse{}, clientErr
	}})
	if _, err := service.Search(t.Context(), "http", Options{}); !errors.Is(err, clientErr) {
		t.Fatalf("Search() error = %v, want wrapped client error", err)
	}
}
