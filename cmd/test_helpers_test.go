package cmd

import (
	"context"

	"github.com/heshanpadmasiri/central-search/internal/catalog"
	"github.com/heshanpadmasiri/central-search/internal/search"
)

type fakeSearchService struct {
	searchFunc func(context.Context, string, search.Options) (search.Response, error)
}

func (f fakeSearchService) Search(ctx context.Context, query string, options search.Options) (search.Response, error) {
	if f.searchFunc == nil {
		panic("unexpected Search call")
	}
	return f.searchFunc(ctx, query, options)
}

type fakeDocumentationService struct {
	documentationFunc func(context.Context, catalog.PackageSelector) (catalog.PackageDocumentation, error)
}

func (f fakeDocumentationService) Documentation(ctx context.Context, selector catalog.PackageSelector) (catalog.PackageDocumentation, error) {
	if f.documentationFunc == nil {
		panic("unexpected Documentation call")
	}
	return f.documentationFunc(ctx, selector)
}
