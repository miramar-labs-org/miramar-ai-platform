// Package templates embeds the deployment templates this binary ships with.
//
// Each template type is its own top-level directory containing a complete
// Helm chart. Adding a new type means adding a new directory here and listing
// it in Available — no other factory code changes.
package templates

import "embed"

//go:embed all:serving-vllm
var FS embed.FS

// Available lists the template types this binary knows how to load or copy.
var Available = []string{"serving-vllm"}
