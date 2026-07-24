package cmd

import (
	"fmt"
	"io"

	skill "github.com/heshanpadmasiri/central-search/central-search-skill"
	"github.com/spf13/cobra"
)

func newLLMCommand(out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "llm",
		Short: "Print usage instructions for language models",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if _, err := io.WriteString(out, skill.Markdown); err != nil {
				return fmt.Errorf("write LLM instructions: %w", err)
			}
			return nil
		},
	}
}
