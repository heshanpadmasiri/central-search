// Package search implements the package-search use case.
package search

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/heshanpadmasiri/central-search/internal/central"
)

var (
	// ErrEmptyQuery indicates that a search query contains no non-whitespace
	// characters.
	ErrEmptyQuery = errors.New("search query must not be empty")
	// ErrInvalidLimit indicates that a search limit is not positive.
	ErrInvalidLimit = errors.New("search limit must be greater than zero")
)

// PackageClient is the Central operation needed to search packages.
type PackageClient interface {
	SearchPackages(context.Context, string, central.SearchPackagesOptions) (central.SearchPackagesResponse, error)
}

// Service searches for packages.
type Service struct {
	client PackageClient
}

// NewService constructs a search service.
func NewService(client PackageClient) *Service {
	return &Service{client: client}
}

// Options controls a search.
type Options struct {
	// Limit is nil when Central should use its default.
	Limit *int
}

// Response contains packages returned by a search.
type Response struct {
	Packages []Package
}

// Package is a package shown in search output.
type Package struct {
	Organization string
	Package      string
	Version      string
	Summary      string
}

// Search searches Central and converts its response into application types.
func (s *Service) Search(ctx context.Context, query string, options Options) (Response, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return Response{}, ErrEmptyQuery
	}
	if options.Limit != nil && *options.Limit <= 0 {
		return Response{}, ErrInvalidLimit
	}

	result, err := s.client.SearchPackages(ctx, query, central.SearchPackagesOptions{Limit: options.Limit})
	if err != nil {
		return Response{}, fmt.Errorf("search Central packages: %w", err)
	}

	packages := make([]Package, len(result.Packages))
	for i, item := range result.Packages {
		packages[i] = Package{
			Organization: item.Organization,
			Package:      item.Name,
			Version:      item.Version,
			Summary:      item.Summary,
		}
	}
	return Response{Packages: packages}, nil
}
