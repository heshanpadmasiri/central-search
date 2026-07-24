package cmd

import (
	"fmt"
	"io"

	"github.com/heshanpadmasiri/central-search/internal/catalog"
	"github.com/spf13/cobra"
)

func newManCommand(service catalog.Service, out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "man <package|organization/package>",
		Short: "Show documentation for the latest package version",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			selector, err := catalog.ParsePackageSelector(args[0])
			if err != nil {
				return err
			}
			documentation, err := service.Documentation(command.Context(), selector)
			if err != nil {
				return fmt.Errorf("get documentation for %s: %w", selector, err)
			}
			if _, err := out.Write(documentation.Content); err != nil {
				return fmt.Errorf("write package documentation: %w", err)
			}
			return nil
		},
	}
}
