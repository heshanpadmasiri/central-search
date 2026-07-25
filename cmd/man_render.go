package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/heshanpadmasiri/central-search/internal/catalog"
	"golang.org/x/term"
)

const defaultTerminalWidth = 100

type fileDescriptor interface{ Fd() uintptr }

func writeHumanDocumentation(ctx context.Context, out, errOut io.Writer, documentation catalog.Package) error {
	markdown := formatDocumentationMarkdown(documentation)
	if !isTerminalWriter(out) || strings.EqualFold(os.Getenv("TERM"), "dumb") {
		if _, err := io.WriteString(out, markdown); err != nil {
			return fmt.Errorf("write package documentation: %w", err)
		}
		return nil
	}

	width := defaultTerminalWidth
	if fd, ok := out.(fileDescriptor); ok {
		if terminalWidth, _, err := term.GetSize(int(fd.Fd())); err == nil && terminalWidth > 0 {
			width = terminalWidth
		}
	}
	rendered, err := renderTerminalMarkdown(markdown, width, os.Getenv("NO_COLOR") != "")
	if err != nil {
		return fmt.Errorf("render package documentation: %w", err)
	}
	return writePagedDocumentation(ctx, out, errOut, rendered)
}

func writePagedDocumentation(ctx context.Context, out, errOut io.Writer, rendered string) error {
	pager := resolvePager()
	if pager == "" {
		if _, err := io.WriteString(out, rendered); err != nil {
			return fmt.Errorf("write package documentation: %w", err)
		}
		return nil
	}
	command := exec.CommandContext(ctx, "sh", "-c", pager)
	command.Stdin = strings.NewReader(rendered)
	command.Stdout = out
	command.Stderr = errOut
	if err := command.Run(); err != nil {
		return fmt.Errorf("run documentation pager %q: %w", pager, err)
	}
	return nil
}

func isTerminalWriter(out io.Writer) bool {
	file, ok := out.(fileDescriptor)
	return ok && term.IsTerminal(int(file.Fd()))
}

func renderTerminalMarkdown(markdown string, width int, noColor bool) (string, error) {
	style := glamour.WithAutoStyle()
	if noColor {
		style = glamour.WithStandardStyle("notty")
	}
	renderer, err := glamour.NewTermRenderer(style, glamour.WithWordWrap(width))
	if err != nil {
		return "", err
	}
	defer renderer.Close()
	return renderer.Render(markdown)
}

func resolvePager() string {
	if pager := strings.TrimSpace(os.Getenv("MANPAGER")); pager != "" {
		return pager
	}
	if pager := strings.TrimSpace(os.Getenv("PAGER")); pager != "" {
		return pager
	}
	if _, err := exec.LookPath("less"); err == nil {
		return "less -R"
	}
	return ""
}

type markdownDocument struct{ strings.Builder }

func formatDocumentationMarkdown(documentation catalog.Package) string {
	var document markdownDocument
	title := strings.Trim(strings.TrimSpace(documentation.Organization)+"/"+strings.TrimSpace(documentation.Name), "/")
	if title == "" {
		title = "Package documentation"
	}
	document.heading(1, title)
	document.label("Version", documentation.Version)
	document.paragraph(documentation.Summary)
	if documentation.Overview != "" {
		document.heading(2, "Package overview")
		document.markdown(documentation.Overview, 3)
	}
	if !documentation.Complete || len(documentation.Warnings) > 0 {
		document.heading(2, "Warnings")
		if !documentation.Complete {
			document.paragraph("Documentation is incomplete.")
		}
		for _, warning := range documentation.Warnings {
			message := strings.TrimSpace(warning.Message)
			if module := strings.TrimSpace(warning.Module); module != "" {
				message = module + ": " + message
			}
			document.listItem(message)
		}
		document.blank()
	}
	for _, module := range documentation.Modules {
		document.module(module)
	}
	return document.String()
}

func (d *markdownDocument) heading(level int, title string) {
	if strings.TrimSpace(title) == "" {
		return
	}
	if level > 6 {
		level = 6
	}
	d.WriteString(strings.Repeat("#", level) + " " + title + "\n\n")
}

func (d *markdownDocument) blank() {
	if d.Len() > 0 && !strings.HasSuffix(d.String(), "\n\n") {
		d.WriteByte('\n')
	}
}

func (d *markdownDocument) paragraph(value string) {
	value = strings.TrimSpace(value)
	if value != "" {
		d.WriteString(value + "\n\n")
	}
}

func (d *markdownDocument) markdown(value string, minimumHeadingLevel int) {
	value = strings.TrimSpace(value)
	if value != "" {
		d.WriteString(demoteMarkdownHeadings(value, minimumHeadingLevel) + "\n\n")
	}
}

func demoteMarkdownHeadings(value string, minimumLevel int) string {
	lines := strings.Split(value, "\n")
	fence := ""
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if fence != "" {
			if strings.HasPrefix(trimmed, fence) {
				fence = ""
			}
			continue
		}
		if strings.HasPrefix(trimmed, "```") {
			fence = "```"
			continue
		}
		if strings.HasPrefix(trimmed, "~~~") {
			fence = "~~~"
			continue
		}
		level := 0
		for level < len(line) && level < 6 && line[level] == '#' {
			level++
		}
		if level == 0 || level >= len(line) || (line[level] != ' ' && line[level] != '\t') || level >= minimumLevel {
			continue
		}
		if minimumLevel > 6 {
			minimumLevel = 6
		}
		lines[i] = strings.Repeat("#", minimumLevel) + line[level:]
	}
	return strings.Join(lines, "\n")
}

func (d *markdownDocument) label(name, value string) {
	if value = strings.TrimSpace(value); value != "" {
		d.WriteString("**" + name + ":** " + value + "\n\n")
	}
}

func (d *markdownDocument) flags(values ...string) {
	present := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			present = append(present, value)
		}
	}
	if len(present) > 0 {
		d.WriteString("**Attributes:** " + strings.Join(present, ", ") + "\n\n")
	}
}

func (d *markdownDocument) signature(value string) {
	if value = strings.TrimSpace(value); value != "" {
		d.WriteString("````ballerina\n" + value + "\n````\n\n")
	}
}

func (d *markdownDocument) listItem(value string) {
	if value = strings.TrimSpace(value); value != "" {
		d.WriteString("- " + value + "\n")
	}
}

func (d *markdownDocument) namedItem(name, signature, documentation string, documentationHeadingLevel int) {
	if name != "" {
		d.WriteString("**`" + name + "`**\n\n")
	}
	d.signature(signature)
	d.markdown(documentation, documentationHeadingLevel)
}

func (d *markdownDocument) symbol(symbol catalog.Symbol, level int) {
	d.heading(level, "`"+symbol.Name+"`")
	d.label("Kind", symbol.Kind)
	d.flags(flag(symbol.Deprecated, "deprecated"), flag(symbol.ReadOnly, "read-only"))
	d.signature(symbol.Signature)
	d.markdown(symbol.Documentation, level+1)
}

func flag(enabled bool, value string) string {
	if enabled {
		return value
	}
	return ""
}

func (d *markdownDocument) module(module catalog.Module) {
	d.heading(2, "Module: "+module.Name)
	d.flags(flag(module.IsDefault, "default module"))
	d.paragraph(module.Summary)
	if module.Overview != "" {
		d.heading(3, "Overview")
		d.markdown(module.Overview, 4)
	}
	if len(module.Functions) > 0 {
		d.heading(3, "Functions")
		for _, item := range module.Functions {
			d.function(item, 4, true)
		}
	}
	if len(module.Resources) > 0 {
		d.heading(3, "Resources")
		for _, item := range module.Resources {
			d.function(item, 4, true)
		}
	}
	if len(module.Classes) > 0 {
		d.heading(3, "Classes")
		for _, item := range module.Classes {
			d.class(item)
		}
	}
	if len(module.Clients) > 0 {
		d.heading(3, "Clients")
		for _, item := range module.Clients {
			d.client(item)
		}
	}
	if len(module.Listeners) > 0 {
		d.heading(3, "Listeners")
		for _, item := range module.Listeners {
			d.listener(item)
		}
	}
	if len(module.Objects) > 0 {
		d.heading(3, "Objects")
		for _, item := range module.Objects {
			d.object(item)
		}
	}
	if len(module.Services) > 0 {
		d.heading(3, "Services")
		for _, item := range module.Services {
			d.service(item)
		}
	}
	if len(module.Records) > 0 {
		d.heading(3, "Records")
		for _, item := range module.Records {
			d.record(item)
		}
	}
	if len(module.Enums) > 0 {
		d.heading(3, "Enums")
		for _, item := range module.Enums {
			d.enum(item)
		}
	}
	if len(module.Types) > 0 {
		d.heading(3, "Types")
		for _, item := range module.Types {
			d.typeDefinition(item)
		}
	}
	if len(module.Errors) > 0 {
		d.heading(3, "Errors")
		for _, item := range module.Errors {
			d.errorType(item)
		}
	}
	if len(module.Constants) > 0 {
		d.heading(3, "Constants")
		for _, item := range module.Constants {
			d.constant(item)
		}
	}
	if len(module.Variables) > 0 {
		d.heading(3, "Variables")
		for _, item := range module.Variables {
			d.variable(item)
		}
	}
	if len(module.Configurables) > 0 {
		d.heading(3, "Configurables")
		for _, item := range module.Configurables {
			d.variable(item)
		}
	}
	if len(module.Annotations) > 0 {
		d.heading(3, "Annotations")
		for _, item := range module.Annotations {
			d.annotation(item)
		}
	}
}

func (d *markdownDocument) function(item catalog.Function, level int, expandSignature bool) {
	d.symbol(item.Symbol, level)
	d.flags(flag(item.Isolated, "isolated"), flag(item.External, "external"), flag(item.Remote, "remote"), flag(item.Resource, "resource"))
	if !expandSignature {
		return
	}
	d.label("Accessor", item.Accessor)
	d.label("Resource path", item.ResourcePath)
	if len(item.Arguments) > 0 {
		d.heading(level+1, "Parameters")
		for _, argument := range item.Arguments {
			d.namedItem(argument.Name, argument.Signature, argument.Documentation, level+2)
			d.label("Type", argument.Type)
			d.label("Default", argument.DefaultValue)
			d.flags(flag(argument.Optional, "optional"), flag(argument.Rest, "rest"))
		}
	}
	if len(item.Returns) > 0 {
		d.heading(level+1, "Returns")
		for _, returned := range item.Returns {
			name := returned.Name
			if name == "" {
				name = returned.Type
			}
			d.namedItem(name, returned.Type, returned.Documentation, level+2)
		}
	}
}

func (d *markdownDocument) fields(items []catalog.Field, level int) {
	if len(items) == 0 {
		return
	}
	d.heading(level, "Fields")
	for _, item := range items {
		d.namedItem(item.Name, item.Signature, item.Documentation, level+1)
		d.label("Kind", item.Kind)
		d.label("Type", item.Type)
		d.label("Default", item.DefaultValue)
		d.flags(flag(item.Deprecated, "deprecated"), flag(item.ReadOnly, "read-only"), flag(item.Optional, "optional"))
	}
}

func (d *markdownDocument) methods(title string, items []catalog.Function, level int) {
	if len(items) == 0 {
		return
	}
	d.heading(level, title)
	for _, item := range items {
		d.function(item, level+1, false)
	}
}

func (d *markdownDocument) class(item catalog.Class) {
	d.symbol(item.Symbol, 4)
	d.flags(flag(item.Isolated, "isolated"))
	d.fields(item.Fields, 5)
	if item.Init != nil {
		d.heading(5, "Initializer")
		d.function(*item.Init, 6, false)
	}
	d.methods("Methods", item.Methods, 5)
}

func (d *markdownDocument) client(item catalog.Client) {
	d.symbol(item.Symbol, 4)
	d.flags(flag(item.Isolated, "isolated"))
	d.fields(item.Fields, 5)
	if item.Init != nil {
		d.heading(5, "Initializer")
		d.function(*item.Init, 6, false)
	}
	d.methods("Methods", item.Methods, 5)
	d.methods("Remote methods", item.RemoteMethods, 5)
	d.methods("Resource methods", item.ResourceMethods, 5)
}

func (d *markdownDocument) listener(item catalog.Listener) {
	d.symbol(item.Symbol, 4)
	d.flags(flag(item.Isolated, "isolated"))
	d.fields(item.Fields, 5)
	if item.Init != nil {
		d.heading(5, "Initializer")
		d.function(*item.Init, 6, false)
	}
	d.methods("Methods", item.Methods, 5)
	d.methods("Lifecycle methods", item.LifecycleMethods, 5)
}

func (d *markdownDocument) includes(values []string) {
	if len(values) == 0 {
		return
	}
	d.heading(5, "Includes")
	for _, value := range values {
		d.listItem("`" + value + "`")
	}
	d.blank()
}

func (d *markdownDocument) object(item catalog.Object) {
	d.symbol(item.Symbol, 4)
	d.flags(flag(item.Distinct, "distinct"))
	d.includes(item.Includes)
	d.fields(item.Fields, 5)
	d.methods("Methods", item.Methods, 5)
}

func (d *markdownDocument) service(item catalog.ServiceDeclaration) {
	d.symbol(item.Symbol, 4)
	d.flags(flag(item.Distinct, "distinct"))
	d.includes(item.Includes)
	d.fields(item.Fields, 5)
	d.methods("Methods", item.Methods, 5)
}

func (d *markdownDocument) record(item catalog.Record) {
	d.symbol(item.Symbol, 4)
	d.flags(flag(item.Closed, "closed"))
	d.includes(item.Includes)
	if len(item.Fields) == 0 {
		return
	}
	d.heading(5, "Fields")
	for _, field := range item.Fields {
		d.namedItem(field.Name, field.Signature, field.Documentation, 6)
		d.label("Kind", field.Kind)
		d.label("Type", field.Type)
		d.label("Default", field.DefaultValue)
		d.flags(flag(field.Deprecated, "deprecated"), flag(field.ReadOnly, "read-only"), flag(field.Optional, "optional"))
	}
}

func (d *markdownDocument) enum(item catalog.Enum) {
	d.symbol(item.Symbol, 4)
	if len(item.Members) == 0 {
		return
	}
	d.heading(5, "Members")
	for _, member := range item.Members {
		d.namedItem(member.Name, member.Signature, member.Documentation, 6)
		d.label("Kind", member.Kind)
		d.label("Value", member.Value)
		d.flags(flag(member.Deprecated, "deprecated"), flag(member.ReadOnly, "read-only"))
	}
}

func (d *markdownDocument) typeDefinition(item catalog.TypeDefinition) {
	d.symbol(item.Symbol, 4)
	d.label("Type kind", item.TypeKind)
	d.label("Definition", item.Definition)
	if len(item.Members) > 0 {
		d.heading(5, "Members")
		for _, member := range item.Members {
			d.listItem("`" + member + "`")
		}
		d.blank()
	}
}

func (d *markdownDocument) errorType(item catalog.ErrorType) {
	d.symbol(item.Symbol, 4)
	d.label("Detail type", item.DetailType)
	d.flags(flag(item.Distinct, "distinct"))
}

func (d *markdownDocument) constant(item catalog.Constant) {
	d.symbol(item.Symbol, 4)
	d.label("Type", item.Type)
	d.label("Value", item.Value)
}

func (d *markdownDocument) variable(item catalog.Variable) {
	d.symbol(item.Symbol, 4)
	d.label("Type", item.Type)
	d.label("Default", item.DefaultValue)
}

func (d *markdownDocument) annotation(item catalog.Annotation) {
	d.symbol(item.Symbol, 4)
	d.label("Type", item.Type)
	if len(item.AttachmentPoints) > 0 {
		d.label("Attachment points", strings.Join(item.AttachmentPoints, ", "))
	}
}
