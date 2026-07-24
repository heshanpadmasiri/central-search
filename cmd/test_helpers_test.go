package cmd

import (
	"context"

	"github.com/heshanpadmasiri/central-search/internal/catalog"
)

type fakeService struct {
	searchFunc        func(context.Context, string) ([]catalog.PackageSummary, error)
	documentationFunc func(context.Context, catalog.PackageSelector) (catalog.PackageDocumentation, error)
}

func (f fakeService) Search(ctx context.Context, query string) ([]catalog.PackageSummary, error) {
	if f.searchFunc == nil {
		panic("unexpected Search call")
	}
	return f.searchFunc(ctx, query)
}

func (f fakeService) Documentation(ctx context.Context, selector catalog.PackageSelector) (catalog.PackageDocumentation, error) {
	if f.documentationFunc == nil {
		panic("unexpected Documentation call")
	}
	return f.documentationFunc(ctx, selector)
}
