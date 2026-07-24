// Package catalog defines package documentation use cases and the stable JSON model.
package catalog

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/heshanpadmasiri/central-search/internal/search"
)

const PackageSchemaVersion = 1

var (
	ErrNoMatches  = errors.New("no packages matched")
	ErrNoVersions = errors.New("matched package has no versions")
)

// PackageSearcher finds packages through Central search.
type PackageSearcher interface {
	Search(context.Context, string, search.Options) (search.Response, error)
}

// VersionClient returns versions for a resolved package.
type VersionClient interface {
	PackageVersions(context.Context, string, string) ([]string, error)
}

// ModuleScraper scrapes one Central module page.
type ModuleScraper interface {
	ScrapeModule(context.Context, string, string, string) (ScrapedModule, error)
}

// Service resolves package queries and assembles all available module docs.
type Service struct {
	searcher PackageSearcher
	versions VersionClient
	scraper  ModuleScraper
}

// NewService constructs the package documentation service.
func NewService(searcher PackageSearcher, versions VersionClient, scraper ModuleScraper) *Service {
	return &Service{searcher: searcher, versions: versions, scraper: scraper}
}

// DocumentationService is the operation used by the CLI.
type DocumentationService interface {
	Documentation(context.Context, string) (Package, error)
}

// PackageMatch identifies one ambiguous search result.
type PackageMatch struct {
	Organization string
	Package      string
}

// AmbiguousPackageError reports every distinct package matching a query.
type AmbiguousPackageError struct {
	Query   string
	Matches []PackageMatch
}

func (e *AmbiguousPackageError) Error() string {
	matches := make([]string, len(e.Matches))
	for i, match := range e.Matches {
		matches[i] = match.Organization + "/" + match.Package
	}
	return fmt.Sprintf("package query %q is ambiguous: %s", e.Query, strings.Join(matches, ", "))
}

// Documentation resolves query and retrieves structured package documentation.
func (s *Service) Documentation(ctx context.Context, query string) (Package, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return Package{}, search.ErrEmptyQuery
	}
	const pageSize = 100
	matches := make([]PackageMatch, 0)
	seen := make(map[string]struct{})
	offset := 0
	for {
		limit, requestedOffset := pageSize, offset
		response, err := s.searcher.Search(ctx, query, search.Options{Limit: &limit, Offset: &requestedOffset})
		if err != nil {
			return Package{}, fmt.Errorf("resolve package query: %w", err)
		}
		for _, item := range response.Packages {
			key := item.Organization + "\x00" + item.Package
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			matches = append(matches, PackageMatch{Organization: item.Organization, Package: item.Package})
		}
		consumed := len(response.Packages)
		if consumed == 0 || response.Offset+consumed >= response.Count {
			break
		}
		next := response.Offset + consumed
		if next <= offset {
			return Package{}, fmt.Errorf("resolve package query: Central search pagination made no progress at offset %d", offset)
		}
		offset = next
	}
	if len(matches) == 0 {
		return Package{}, fmt.Errorf("%w: %q", ErrNoMatches, query)
	}
	if len(matches) > 1 {
		for _, candidate := range matches {
			if candidate.Organization+"/"+candidate.Package == query {
				matches = []PackageMatch{candidate}
				break
			}
		}
	}
	if len(matches) > 1 {
		return Package{}, &AmbiguousPackageError{Query: query, Matches: matches}
	}
	match := matches[0]
	versions, err := s.versions.PackageVersions(ctx, match.Organization, match.Package)
	if err != nil {
		return Package{}, fmt.Errorf("get versions for %s/%s: %w", match.Organization, match.Package, err)
	}
	if len(versions) == 0 {
		return Package{}, fmt.Errorf("%w: %s/%s", ErrNoVersions, match.Organization, match.Package)
	}
	version := versions[0]
	defaultResult, err := s.scraper.ScrapeModule(ctx, match.Organization, match.Package, version)
	if err != nil {
		return Package{}, fmt.Errorf("scrape default module %s: %w", match.Package, err)
	}
	result := Package{
		SchemaVersion: PackageSchemaVersion,
		Complete:      true,
		Warnings:      []Warning{},
		Organization:  match.Organization,
		Name:          match.Package,
		Version:       version,
		Summary:       defaultResult.PackageSummary,
		Overview:      defaultResult.PackageOverview,
		Modules:       []Module{defaultResult.Module},
	}
	seenModules := map[string]struct{}{defaultResult.Module.Name: {}}
	for _, related := range defaultResult.RelatedModules {
		if related.IsDefault {
			continue
		}
		if _, exists := seenModules[related.Name]; exists {
			continue
		}
		seenModules[related.Name] = struct{}{}
		moduleResult, scrapeErr := s.scraper.ScrapeModule(ctx, match.Organization, related.Name, version)
		if scrapeErr != nil {
			if errors.Is(scrapeErr, context.Canceled) || errors.Is(scrapeErr, context.DeadlineExceeded) {
				return Package{}, fmt.Errorf("scrape related module %s: %w", related.Name, scrapeErr)
			}
			result.Complete = false
			result.Warnings = append(result.Warnings, Warning{Module: related.Name, Message: scrapeErr.Error()})
			continue
		}
		result.Modules = append(result.Modules, moduleResult.Module)
	}
	if len(result.Modules) == 0 {
		return Package{}, errors.New("no module documentation was produced")
	}
	return result, nil
}

// Package is the stable structured documentation emitted by man --json.
type Package struct {
	SchemaVersion int       `json:"schemaVersion"`
	Complete      bool      `json:"complete"`
	Warnings      []Warning `json:"warnings"`
	Organization  string    `json:"organization"`
	Name          string    `json:"name"`
	Version       string    `json:"version"`
	Summary       string    `json:"summary"`
	Overview      string    `json:"overview"`
	Modules       []Module  `json:"modules"`
}

// Warning describes an omitted related module.
type Warning struct {
	Module  string `json:"module"`
	Message string `json:"message"`
}

// ScrapedModule is the result needed to assemble a package.
type ScrapedModule struct {
	PackageSummary  string
	PackageOverview string
	Module          Module
	RelatedModules  []ModuleReference
}

// ModuleReference identifies another module in the same package version.
type ModuleReference struct {
	Name      string
	IsDefault bool
}

// Symbol contains fields shared by declarations and declaration members.
type Symbol struct {
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	Signature     string `json:"signature"`
	Documentation string `json:"documentation"`
	Deprecated    bool   `json:"deprecated"`
	ReadOnly      bool   `json:"readOnly"`
}

type Module struct {
	Name          string               `json:"name"`
	Summary       string               `json:"summary"`
	Overview      string               `json:"overview"`
	IsDefault     bool                 `json:"isDefault"`
	Functions     []Function           `json:"functions"`
	Resources     []Function           `json:"resources"`
	Classes       []Class              `json:"classes"`
	Clients       []Client             `json:"clients"`
	Listeners     []Listener           `json:"listeners"`
	Objects       []Object             `json:"objects"`
	Services      []ServiceDeclaration `json:"services"`
	Records       []Record             `json:"records"`
	Enums         []Enum               `json:"enums"`
	Types         []TypeDefinition     `json:"types"`
	Errors        []ErrorType          `json:"errors"`
	Constants     []Constant           `json:"constants"`
	Variables     []Variable           `json:"variables"`
	Configurables []Variable           `json:"configurables"`
	Annotations   []Annotation         `json:"annotations"`
}

type Function struct {
	Symbol
	Arguments    []Argument    `json:"arguments"`
	Returns      []ReturnValue `json:"returns"`
	Accessor     string        `json:"accessor,omitempty"`
	ResourcePath string        `json:"resourcePath,omitempty"`
	Isolated     bool          `json:"isolated"`
	External     bool          `json:"external"`
	Remote       bool          `json:"remote"`
	Resource     bool          `json:"resource"`
}
type Argument struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Signature     string `json:"signature"`
	Documentation string `json:"documentation"`
	DefaultValue  string `json:"defaultValue,omitempty"`
	Optional      bool   `json:"optional"`
	Rest          bool   `json:"rest"`
}
type ReturnValue struct {
	Name          string `json:"name,omitempty"`
	Type          string `json:"type"`
	Documentation string `json:"documentation"`
}
type Field struct {
	Symbol
	Type         string `json:"type"`
	Optional     bool   `json:"optional"`
	DefaultValue string `json:"defaultValue,omitempty"`
}
type Record struct {
	Symbol
	Closed   bool          `json:"closed"`
	Includes []string      `json:"includes"`
	Fields   []RecordField `json:"fields"`
}
type RecordField struct {
	Symbol
	Type         string `json:"type"`
	Optional     bool   `json:"optional"`
	DefaultValue string `json:"defaultValue,omitempty"`
}
type Class struct {
	Symbol
	Fields   []Field    `json:"fields"`
	Init     *Function  `json:"init,omitempty"`
	Methods  []Function `json:"methods"`
	Isolated bool       `json:"isolated"`
}
type Client struct {
	Symbol
	Fields          []Field    `json:"fields"`
	Init            *Function  `json:"init,omitempty"`
	Methods         []Function `json:"methods"`
	RemoteMethods   []Function `json:"remoteMethods"`
	ResourceMethods []Function `json:"resourceMethods"`
	Isolated        bool       `json:"isolated"`
}
type Listener struct {
	Symbol
	Fields           []Field    `json:"fields"`
	Init             *Function  `json:"init,omitempty"`
	Methods          []Function `json:"methods"`
	LifecycleMethods []Function `json:"lifecycleMethods"`
	Isolated         bool       `json:"isolated"`
}
type Object struct {
	Symbol
	Fields   []Field    `json:"fields"`
	Methods  []Function `json:"methods"`
	Includes []string   `json:"includes"`
	Distinct bool       `json:"distinct"`
}

// ServiceDeclaration is a service declaration (named to avoid colliding with the orchestration Service).
type ServiceDeclaration struct {
	Symbol
	Fields   []Field    `json:"fields"`
	Methods  []Function `json:"methods"`
	Includes []string   `json:"includes"`
	Distinct bool       `json:"distinct"`
}
type TypeDefinition struct {
	Symbol
	TypeKind   string   `json:"typeKind"`
	Definition string   `json:"definition"`
	Members    []string `json:"members"`
}
type Enum struct {
	Symbol
	Members []EnumMember `json:"members"`
}
type EnumMember struct {
	Symbol
	Value string `json:"value,omitempty"`
}
type ErrorType struct {
	Symbol
	DetailType string `json:"detailType"`
	Distinct   bool   `json:"distinct"`
}
type Constant struct {
	Symbol
	Type  string `json:"type"`
	Value string `json:"value"`
}
type Variable struct {
	Symbol
	Type         string `json:"type"`
	DefaultValue string `json:"defaultValue,omitempty"`
}
type Annotation struct {
	Symbol
	Type             string   `json:"type"`
	AttachmentPoints []string `json:"attachmentPoints"`
}
