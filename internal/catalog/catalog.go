// Package catalog defines the Ballerina Central operations used by the CLI.
package catalog

import (
	"context"
	"fmt"
	"strings"
	"unicode"
)

// PackageSelector identifies a package. Organization is empty when the user
// supplied an unqualified package name.
type PackageSelector struct {
	Organization string
	Package      string
}

// ParsePackageSelector parses package or organization/package.
func ParsePackageSelector(value string) (PackageSelector, error) {
	value = strings.TrimSpace(value)
	parts := strings.Split(value, "/")
	if len(parts) < 1 || len(parts) > 2 {
		return PackageSelector{}, fmt.Errorf("invalid package %q: expected package or organization/package", value)
	}
	for _, part := range parts {
		if part == "" || strings.IndexFunc(part, unicode.IsSpace) >= 0 {
			return PackageSelector{}, fmt.Errorf("invalid package %q: expected package or organization/package", value)
		}
	}
	if len(parts) == 1 {
		return PackageSelector{Package: parts[0]}, nil
	}
	return PackageSelector{Organization: parts[0], Package: parts[1]}, nil
}

// String returns the selector in command-line form.
func (s PackageSelector) String() string {
	if s.Organization == "" {
		return s.Package
	}
	return s.Organization + "/" + s.Package
}

// PackageSummary is a package returned by a search.
type PackageSummary struct {
	Organization string
	Package      string
	Summary      string
}

// PackageDocumentation contains the opaque documentation response for the
// latest version of a resolved package. Parsing and rendering are deliberately
// deferred until the Central API response format is integrated.
type PackageDocumentation struct {
	Organization string
	Package      string
	Version      string
	ContentType  string
	Content      []byte
}

// Service provides access to Ballerina Central. Search implementations must
// match query case-insensitively as a substring of either the organization or
// package name, but not the summary.
type Service interface {
	Search(ctx context.Context, query string) ([]PackageSummary, error)
	Documentation(ctx context.Context, selector PackageSelector) (PackageDocumentation, error)
}
