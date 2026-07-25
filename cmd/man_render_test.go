package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/heshanpadmasiri/central-search/internal/catalog"
)

func TestFormatDocumentationMarkdownRendersCompleteModel(t *testing.T) {
	base := catalog.Symbol{Name: "Thing", Signature: "public type Thing record {}", Documentation: "Thing docs", Deprecated: true, ReadOnly: true}
	function := catalog.Function{
		Symbol:    catalog.Symbol{Name: "call", Signature: "remote function call(string value) returns string", Documentation: "Call docs"},
		Arguments: []catalog.Argument{{Name: "value", Type: "string", Signature: "string value = x", Documentation: "Value docs", DefaultValue: "x", Optional: true, Rest: true}},
		Returns:   []catalog.ReturnValue{{Name: "result", Type: "string", Documentation: "Result docs"}},
		Accessor:  "get", ResourcePath: "/items", Isolated: true, External: true, Remote: true, Resource: true,
	}
	field := catalog.Field{Symbol: catalog.Symbol{Name: "field", Signature: "string field", Documentation: "Field docs", Deprecated: true, ReadOnly: true}, Type: "string", Optional: true, DefaultValue: "value"}
	module := catalog.Module{
		Name: "sample", Summary: "Module summary", Overview: "Module overview", IsDefault: true,
		Functions: []catalog.Function{function}, Resources: []catalog.Function{function},
		Classes:       []catalog.Class{{Symbol: base, Fields: []catalog.Field{field}, Init: &function, Methods: []catalog.Function{function}, Isolated: true}},
		Clients:       []catalog.Client{{Symbol: base, Fields: []catalog.Field{field}, Init: &function, Methods: []catalog.Function{function}, RemoteMethods: []catalog.Function{function}, ResourceMethods: []catalog.Function{function}, Isolated: true}},
		Listeners:     []catalog.Listener{{Symbol: base, Fields: []catalog.Field{field}, Init: &function, Methods: []catalog.Function{function}, LifecycleMethods: []catalog.Function{function}, Isolated: true}},
		Objects:       []catalog.Object{{Symbol: base, Fields: []catalog.Field{field}, Methods: []catalog.Function{function}, Includes: []string{"Included"}, Distinct: true}},
		Services:      []catalog.ServiceDeclaration{{Symbol: base, Fields: []catalog.Field{field}, Methods: []catalog.Function{function}, Includes: []string{"Included"}, Distinct: true}},
		Records:       []catalog.Record{{Symbol: base, Closed: true, Includes: []string{"Included"}, Fields: []catalog.RecordField{{Symbol: field.Symbol, Type: field.Type, Optional: true, DefaultValue: field.DefaultValue}}}},
		Enums:         []catalog.Enum{{Symbol: base, Members: []catalog.EnumMember{{Symbol: catalog.Symbol{Name: "MEMBER", Signature: "MEMBER = 1", Documentation: "Member docs", Deprecated: true}, Value: "1"}}}},
		Types:         []catalog.TypeDefinition{{Symbol: base, TypeKind: "union", Definition: "string|int", Members: []string{"string", "int"}}},
		Errors:        []catalog.ErrorType{{Symbol: base, DetailType: "Detail", Distinct: true}},
		Constants:     []catalog.Constant{{Symbol: base, Type: "string", Value: "value"}},
		Variables:     []catalog.Variable{{Symbol: base, Type: "string", DefaultValue: "value"}},
		Configurables: []catalog.Variable{{Symbol: base, Type: "int", DefaultValue: "1"}},
		Annotations:   []catalog.Annotation{{Symbol: base, Type: "Info", AttachmentPoints: []string{"function", "service"}}},
	}
	documentation := catalog.Package{
		Complete: false, Organization: "acme", Name: "sample", Version: "1.2.3", Summary: "Package summary", Overview: "## Existing Markdown",
		Warnings: []catalog.Warning{{Module: "sample.extra", Message: "unavailable"}}, Modules: []catalog.Module{module},
	}

	output := formatDocumentationMarkdown(documentation)
	for _, expected := range []string{
		"# acme/sample", "**Version:** 1.2.3", "## Package overview", "## Existing Markdown", "## Warnings", "sample.extra: unavailable",
		"## Module: sample", "### Overview", "### Functions", "### Resources", "### Classes", "### Clients", "### Listeners", "### Objects", "### Services", "### Records", "### Enums", "### Types", "### Errors", "### Constants", "### Variables", "### Configurables", "### Annotations",
		"##### Parameters", "**Default:** x", "optional, rest", "##### Returns", "**Accessor:** get", "**Resource path:** /items", "isolated, external, remote, resource",
		"##### Initializer", "##### Remote methods", "##### Resource methods", "##### Lifecycle methods", "##### Includes", "`Included`", "closed", "distinct", "**Type kind:** union", "**Definition:** string|int", "**Detail type:** Detail", "**Value:** value", "**Attachment points:** function, service", "deprecated, read-only",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("output does not contain %q", expected)
		}
	}
}

func TestFormatDocumentationMarkdownDoesNotExpandMethods(t *testing.T) {
	method := catalog.Function{
		Symbol:    catalog.Symbol{Kind: "remoteMethod", Name: "call", Signature: "remote function call(string value) returns string", Documentation: "Method docs"},
		Arguments: []catalog.Argument{{Name: "value", Type: "string", Signature: "string value", Documentation: "Value docs"}},
		Returns:   []catalog.ReturnValue{{Type: "string", Documentation: "Return docs"}},
		Remote:    true,
	}
	output := formatDocumentationMarkdown(catalog.Package{
		Complete: true, Organization: "acme", Name: "sample",
		Modules: []catalog.Module{{Name: "sample", Clients: []catalog.Client{{Symbol: catalog.Symbol{Name: "Client"}, RemoteMethods: []catalog.Function{method}}}}},
	})
	for _, expected := range []string{"remote function call(string value) returns string", "Method docs", "**Attributes:** remote"} {
		if !strings.Contains(output, expected) {
			t.Errorf("output does not contain %q:\n%s", expected, output)
		}
	}
	for _, unwanted := range []string{"###### Parameters", "###### Returns", "Value docs", "Return docs"} {
		if strings.Contains(output, unwanted) {
			t.Errorf("output unexpectedly contains %q:\n%s", unwanted, output)
		}
	}
}

func TestDemoteMarkdownHeadingsPreservesFencedCode(t *testing.T) {
	input := "## Section\n\n```markdown\n## Code heading\n```\n\n#### Detail"
	want := "#### Section\n\n```markdown\n## Code heading\n```\n\n#### Detail"
	if got := demoteMarkdownHeadings(input, 4); got != want {
		t.Fatalf("demoteMarkdownHeadings()=\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatDocumentationMarkdownOmitsEmptyParts(t *testing.T) {
	output := formatDocumentationMarkdown(catalog.Package{Complete: true, Organization: "acme", Name: "empty", Modules: []catalog.Module{{Name: "empty"}}})
	for _, unwanted := range []string{"Version:", "Package overview", "Warnings", "### Functions", "Attributes:"} {
		if strings.Contains(output, unwanted) {
			t.Errorf("output unexpectedly contains %q:\n%s", unwanted, output)
		}
	}
	if output != "# acme/empty\n\n## Module: empty\n\n" {
		t.Fatalf("output=%q", output)
	}
}

func TestRenderTerminalMarkdownWithoutColor(t *testing.T) {
	output, err := renderTerminalMarkdown("# Heading\n\nText\n", 80, true)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains([]byte(output), []byte("\x1b[")) {
		t.Fatalf("output contains ANSI color: %q", output)
	}
	if !strings.Contains(output, "Heading") || !strings.Contains(output, "Text") {
		t.Fatalf("output=%q", output)
	}
}

func TestResolvePagerPriority(t *testing.T) {
	t.Setenv("MANPAGER", "most")
	t.Setenv("PAGER", "less")
	if got := resolvePager(); got != "most" {
		t.Fatalf("resolvePager()=%q", got)
	}
	t.Setenv("MANPAGER", "")
	if got := resolvePager(); got != "less" {
		t.Fatalf("resolvePager()=%q", got)
	}
}

func TestWritePagedDocumentationUsesPager(t *testing.T) {
	t.Setenv("MANPAGER", "cat")
	var out bytes.Buffer
	if err := writePagedDocumentation(t.Context(), &out, &bytes.Buffer{}, "rendered documentation\n"); err != nil {
		t.Fatal(err)
	}
	if out.String() != "rendered documentation\n" {
		t.Fatalf("output=%q", out.String())
	}
}
