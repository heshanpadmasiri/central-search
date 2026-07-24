package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/heshanpadmasiri/central-search/internal/catalog"
	"github.com/heshanpadmasiri/central-search/internal/search"
	"github.com/spf13/cobra"
)

var ErrManTextRenderingUnavailable = errors.New("text documentation rendering is not implemented; use --json")

type manOptions struct{ json bool }

func newManCommand(service catalog.DocumentationService, out, errOut io.Writer) *cobra.Command {
	options := manOptions{}
	command := &cobra.Command{
		Use: "man <query>", Short: "Show structured documentation for a uniquely matched package", Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			query := strings.TrimSpace(args[0])
			if query == "" {
				return search.ErrEmptyQuery
			}
			if !options.json {
				return ErrManTextRenderingUnavailable
			}
			documentation, err := service.Documentation(command.Context(), query)
			if err != nil {
				return fmt.Errorf("get documentation for %q: %w", query, err)
			}
			if err := json.NewEncoder(out).Encode(documentation); err != nil {
				return fmt.Errorf("write package documentation JSON: %w", err)
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
