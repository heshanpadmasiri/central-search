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
	// ErrInvalidOffset indicates that a search offset is negative.
	ErrInvalidOffset = errors.New("search offset must not be negative")
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
	// Offset is nil when Central should use its default.
	Offset *int
}

// Response contains packages and Central pagination metadata.
type Response struct {
	Packages []Package
	Count    int
	Offset   int
	Limit    int
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
	if options.Offset != nil && *options.Offset < 0 {
		return Response{}, ErrInvalidOffset
	}

	result, err := s.client.SearchPackages(ctx, query, central.SearchPackagesOptions{Limit: options.Limit, Offset: options.Offset})
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
	return Response{
		Packages: packages,
		Count:    result.Count,
		Offset:   result.Offset,
		Limit:    result.Limit,
	}, nil
}
