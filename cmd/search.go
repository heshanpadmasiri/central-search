package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/heshanpadmasiri/central-search/internal/catalog"
	"github.com/spf13/cobra"
)

// ErrNoMatches indicates that a search completed without finding a package.
var ErrNoMatches = errors.New("no packages matched")

type searchOptions struct {
	json bool
}

type jsonSearchResult struct {
	Organization string `json:"organization"`
	Package      string `json:"package"`
	Summary      string `json:"summary"`
}

func newSearchCommand(service catalog.Service, out io.Writer) *cobra.Command {
	options := searchOptions{}
	command := &cobra.Command{
		Use:   "search <query>",
		Short: "Search package organizations and names",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			query := strings.TrimSpace(args[0])
			if query == "" {
				return errors.New("search query must not be empty")
			}

			results, err := service.Search(command.Context(), query)
			if err != nil {
				return fmt.Errorf("search packages: %w", err)
			}
			sortSearchResults(results)

			if len(results) == 0 {
				if options.json {
					if _, err := io.WriteString(out, "[]\n"); err != nil {
						return fmt.Errorf("write search results: %w", err)
					}
				}
				return ErrNoMatches
			}

			if options.json {
				return writeJSONSearchResults(out, results)
			}
			return writeTextSearchResults(out, results)
		},
	}
	command.Flags().BoolVar(&options.json, "json", false, "output results as JSON")
	return command
}

func sortSearchResults(results []catalog.PackageSummary) {
	sort.SliceStable(results, func(i, j int) bool {
		leftOrg := strings.ToLower(results[i].Organization)
		rightOrg := strings.ToLower(results[j].Organization)
		if leftOrg != rightOrg {
			return leftOrg < rightOrg
		}
		leftPackage := strings.ToLower(results[i].Package)
		rightPackage := strings.ToLower(results[j].Package)
		if leftPackage != rightPackage {
			return leftPackage < rightPackage
		}
		if results[i].Organization != results[j].Organization {
			return results[i].Organization < results[j].Organization
		}
		return results[i].Package < results[j].Package
	})
}

func writeJSONSearchResults(out io.Writer, results []catalog.PackageSummary) error {
	jsonResults := make([]jsonSearchResult, len(results))
	for i, result := range results {
		jsonResults[i] = jsonSearchResult{
			Organization: result.Organization,
			Package:      result.Package,
			Summary:      result.Summary,
		}
	}
	if err := json.NewEncoder(out).Encode(jsonResults); err != nil {
		return fmt.Errorf("write search results: %w", err)
	}
	return nil
}

func writeTextSearchResults(out io.Writer, results []catalog.PackageSummary) error {
	writer := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	for _, result := range results {
		if _, err := fmt.Fprintf(writer, "%s/%s\t%s\n", result.Organization, result.Package, result.Summary); err != nil {
			return fmt.Errorf("write search results: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("write search results: %w", err)
	}
	return nil
}
