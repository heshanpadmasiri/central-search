package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelpListsCommands(t *testing.T) {
	var out bytes.Buffer
	root := NewRootCommand(fakeSearchService{}, fakeDocumentationService{}, IOStreams{Out: &out, ErrOut: &bytes.Buffer{}})
	root.SetArgs([]string{"--help"})
	if err := root.ExecuteContext(t.Context()); err != nil {
		t.Fatalf("ExecuteContext() error = %v", err)
	}
	for _, command := range []string{"search", "man", "llm"} {
		if !strings.Contains(out.String(), command) {
			t.Fatalf("help output does not contain %q: %q", command, out.String())
		}
	}
	if strings.Contains(out.String(), "completion") {
		t.Fatalf("help unexpectedly contains completion command: %q", out.String())
	}
}
