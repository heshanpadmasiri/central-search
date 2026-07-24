package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/heshanpadmasiri/central-search/internal/search"
	"github.com/spf13/cobra"
)

// ErrNoMatches indicates that a search completed without finding a package.
var ErrNoMatches = errors.New("no packages matched")

// SearchService is the package-search operation used by the CLI.
type SearchService interface {
	Search(context.Context, string, search.Options) (search.Response, error)
}

type searchOptions struct {
	json  bool
	limit int
}

type jsonSearchResult struct {
	Organization string `json:"organization"`
	Package      string `json:"package"`
	Version      string `json:"version"`
	Summary      string `json:"summary"`
}

func newSearchCommand(service SearchService, out io.Writer) *cobra.Command {
	options := searchOptions{}
	command := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for packages in Ballerina Central",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			query := strings.TrimSpace(args[0])
			if query == "" {
				return search.ErrEmptyQuery
			}

			searchOptions := search.Options{}
			if command.Flags().Changed("limit") {
				if options.limit <= 0 {
					return search.ErrInvalidLimit
				}
				searchOptions.Limit = &options.limit
			}
			response, err := service.Search(command.Context(), query, searchOptions)
			if err != nil {
				return fmt.Errorf("search packages: %w", err)
			}

			if len(response.Packages) == 0 {
				if options.json {
					if _, err := io.WriteString(out, "[]\n"); err != nil {
						return fmt.Errorf("write search results: %w", err)
					}
				}
				return ErrNoMatches
			}

			if options.json {
				return writeJSONSearchResults(out, response.Packages)
			}
			return writeTextSearchResults(out, response.Packages)
		},
	}
	command.Flags().BoolVar(&options.json, "json", false, "output results as JSON")
	command.Flags().IntVar(&options.limit, "limit", 0, "maximum number of results")
	return command
}

func writeJSONSearchResults(out io.Writer, results []search.Package) error {
	jsonResults := make([]jsonSearchResult, len(results))
	for i, result := range results {
		jsonResults[i] = jsonSearchResult{
			Organization: result.Organization,
			Package:      result.Package,
			Version:      result.Version,
			Summary:      result.Summary,
		}
	}
	if err := json.NewEncoder(out).Encode(jsonResults); err != nil {
		return fmt.Errorf("write search results: %w", err)
	}
	return nil
}

func writeTextSearchResults(out io.Writer, results []search.Package) error {
	writer := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	for _, result := range results {
		if _, err := fmt.Fprintf(writer, "%s/%s\t%s\t%s\n", result.Organization, result.Package, result.Version, result.Summary); err != nil {
			return fmt.Errorf("write search results: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("write search results: %w", err)
	}
	return nil
}
