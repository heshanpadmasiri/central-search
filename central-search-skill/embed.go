// Package skill provides the embedded central-search skill instructions.
package skill

import _ "embed"

// Markdown contains the central-search skill instructions.
//
//go:embed SKILL.md
var Markdown string
