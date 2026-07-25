package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/heshanpadmasiri/central-search/internal/catalog"
	"github.com/heshanpadmasiri/central-search/internal/search"
	"github.com/spf13/cobra"
)

type manOptions struct{ json bool }

func newManCommand(service catalog.DocumentationService, out, errOut io.Writer) *cobra.Command {
	options := manOptions{}
	command := &cobra.Command{
		Use: "man <query>", Short: "Show documentation for a uniquely matched package", Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			query := strings.TrimSpace(args[0])
			if query == "" {
				return search.ErrEmptyQuery
			}
			documentation, err := service.Documentation(command.Context(), query)
			if err != nil {
				return fmt.Errorf("get documentation for %q: %w", query, err)
			}
			if options.json {
				if err := json.NewEncoder(out).Encode(documentation); err != nil {
					return fmt.Errorf("write package documentation JSON: %w", err)
				}
			} else if err := writeHumanDocumentation(command.Context(), out, errOut, documentation); err != nil {
				return err
			}
			for _, warning := range documentation.Warnings {
				if _, err := fmt.Fprintf(errOut, "Warning: %s: %s\n", warning.Module, warning.Message); err != nil {
					return fmt.Errorf("write package documentation warning: %w", err)
				}
			}
			return nil
		},
	}
	command.Flags().BoolVar(&options.json, "json", false, "output structured documentation as JSON")
	return command
}
