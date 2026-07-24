// Package cmd defines the central-search command-line interface.
package cmd

import (
	"io"

	"github.com/heshanpadmasiri/central-search/internal/catalog"
	"github.com/spf13/cobra"
)

// IOStreams contains the standard streams used by the CLI.
type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

// NewRootCommand constructs the central-search root command.
func NewRootCommand(service catalog.Service, streams IOStreams) *cobra.Command {
	command := &cobra.Command{
		Use:           "central-search",
		Short:         "Search and read documentation from Ballerina Central",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	command.SetIn(streams.In)
	command.SetOut(streams.Out)
	command.SetErr(streams.ErrOut)
	command.AddCommand(newSearchCommand(service, streams.Out))
	command.AddCommand(newManCommand(service, streams.Out))
	return command
}
