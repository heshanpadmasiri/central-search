package cmd

import (
	"bytes"
	"testing"

	skill "github.com/heshanpadmasiri/central-search/central-search-skill"
)

func TestLLMPrintsSkillMarkdown(t *testing.T) {
	var out bytes.Buffer
	root := NewRootCommand(fakeSearchService{}, fakeDocumentationService{}, IOStreams{Out: &out, ErrOut: &bytes.Buffer{}})
	root.SetArgs([]string{"llm"})

	if err := root.ExecuteContext(t.Context()); err != nil {
		t.Fatalf("ExecuteContext() error = %v", err)
	}
	if out.String() != skill.Markdown {
		t.Fatalf("output does not match embedded skill markdown")
	}
}
