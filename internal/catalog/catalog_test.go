package catalog

import (
	"context"
	"errors"
	"testing"

	"github.com/heshanpadmasiri/central-search/internal/search"
)

type fakeSearcher func(context.Context, string, search.Options) (search.Response, error)

func (f fakeSearcher) Search(ctx context.Context, q string, o search.Options) (search.Response, error) {
	return f(ctx, q, o)
}

type fakeVersions func(context.Context, string, string) ([]string, error)

func (f fakeVersions) PackageVersions(ctx context.Context, o, p string) ([]string, error) {
	return f(ctx, o, p)
}

type fakeScraper func(context.Context, string, string, string) (ScrapedModule, error)

func (f fakeScraper) ScrapeModule(ctx context.Context, o, m, v string) (ScrapedModule, error) {
	return f(ctx, o, m, v)
}

func TestDocumentationAssemblesPartialPackage(t *testing.T) {
	searcher := fakeSearcher(func(_ context.Context, q string, o search.Options) (search.Response, error) {
		if q != "http client" {
			t.Fatalf("query=%q", q)
		}
		return search.Response{Packages: []search.Package{{Organization: "ballerina", Package: "http"}}, Count: 1, Offset: 0, Limit: 100}, nil
	})
	versions := fakeVersions(func(_ context.Context, o, p string) ([]string, error) { return []string{"2.0.0", "1.0.0"}, nil })
	scraper := fakeScraper(func(_ context.Context, _ string, m, v string) (ScrapedModule, error) {
		if v != "2.0.0" {
			t.Fatalf("version=%q", v)
		}
		if m == "http.bad" {
			return ScrapedModule{}, errors.New("unavailable")
		}
		if m == "http" {
			return ScrapedModule{PackageSummary: "summary", PackageOverview: "overview", Module: Module{Name: "http"}, RelatedModules: []ModuleReference{{Name: "http", IsDefault: true}, {Name: "http.extra"}, {Name: "http.extra"}, {Name: "http.bad"}}}, nil
		}
		return ScrapedModule{Module: Module{Name: m}}, nil
	})
	result, err := NewService(searcher, versions, scraper).Documentation(t.Context(), "  http client ")
	if err != nil {
		t.Fatal(err)
	}
	if result.Complete || len(result.Warnings) != 1 || len(result.Modules) != 2 || result.Version != "2.0.0" {
		t.Fatalf("result=%#v", result)
	}
}

func TestDocumentationPrefersExactQualifiedMatch(t *testing.T) {
	searcher := fakeSearcher(func(_ context.Context, _ string, _ search.Options) (search.Response, error) {
		return search.Response{Packages: []search.Package{
			{Organization: "ballerinax", Package: "postgresql"},
			{Organization: "ballerinax", Package: "postgresql.driver"},
			{Organization: "ballerinax", Package: "postgresql.cdc.driver"},
		}, Count: 3}, nil
	})
	versions := fakeVersions(func(_ context.Context, organization, packageName string) ([]string, error) {
		if organization != "ballerinax" || packageName != "postgresql" {
			t.Fatalf("resolved %s/%s", organization, packageName)
		}
		return []string{"1.0.0"}, nil
	})
	scraper := fakeScraper(func(_ context.Context, _, module, _ string) (ScrapedModule, error) {
		return ScrapedModule{Module: Module{Name: module}}, nil
	})
	result, err := NewService(searcher, versions, scraper).Documentation(t.Context(), "ballerinax/postgresql")
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "postgresql" {
		t.Fatalf("name = %q", result.Name)
	}
}

func TestDocumentationResolutionErrors(t *testing.T) {
	versions := fakeVersions(func(context.Context, string, string) ([]string, error) { return nil, nil })
	scraper := fakeScraper(func(context.Context, string, string, string) (ScrapedModule, error) { panic("unexpected") })
	for _, test := range []struct {
		name     string
		response search.Response
		target   error
	}{
		{"none", search.Response{}, ErrNoMatches},
		{"ambiguous", search.Response{Packages: []search.Package{{Organization: "a", Package: "one"}, {Organization: "b", Package: "two"}}, Count: 2}, &AmbiguousPackageError{}},
		{"versions", search.Response{Packages: []search.Package{{Organization: "a", Package: "one"}}, Count: 1}, ErrNoVersions},
	} {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(fakeSearcher(func(context.Context, string, search.Options) (search.Response, error) { return test.response, nil }), versions, scraper)
			_, err := service.Documentation(t.Context(), "q")
			if test.name == "ambiguous" {
				var target *AmbiguousPackageError
				if !errors.As(err, &target) || len(target.Matches) != 2 {
					t.Fatalf("error=%v", err)
				}
			} else if !errors.Is(err, test.target) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}
