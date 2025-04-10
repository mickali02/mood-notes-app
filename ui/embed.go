// ui/embed.go
package ui

import "embed"

// Embed the html templates and the entire static directory and its contents.
// NOTE: Paths are relative to this ui directory.
//go:embed html static
var Files embed.FS