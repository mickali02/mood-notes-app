// cmd/web/templates.go
package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs" // Required for embed
	"path/filepath"
	"strings"
	"time"

	"github.com/mickali02/mood-notes-app/ui" // Import the ui package with embedded files
	// internal/data import might be removed if no template funcs need it directly
)

// Define a template function map
var functions = template.FuncMap{
	"humanDate": humanDate,
	"safeHTML": func(s string) template.HTML {
		return template.HTML(s)
	},
	// Add more functions if needed
}

// humanDate formats a time.Time object nicely for display.
func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	// Example format: "Monday, Jan 02, 2006 at 03:04 PM" (Uses local time)
	return t.Local().Format("Monday, Jan 02, 2006 at 03:04 PM")
}

// newTemplateCache parses all template files (*.tmpl) from the embedded filesystem (ui.Files)
// using the simpler naming convention and stores them in a map.
func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// 1. Find all the 'page' templates using the new simpler pattern
	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl") // CHANGED pattern
	if err != nil {
		return nil, fmt.Errorf("error globbing pages: %w", err)
	}
	if len(pages) == 0 {
		return nil, errors.New("no page templates (*.tmpl) found in ui/html/pages")
	}

	// 2. Loop through each page template found
	for _, page := range pages {
		name := filepath.Base(page) // Get the filename (e.g., "home.tmpl")

		// 3. Create a new template set for this page, add functions.
		// Start by parsing the specific page file itself.
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, page)
		if err != nil {
			return nil, fmt.Errorf("error parsing page %s: %w", name, err)
		}

		// 4. Parse the base layout template into the set, using the simpler name.
		ts, err = ts.ParseFS(ui.Files, "html/layouts/base.tmpl") // CHANGED path
		if err != nil {
			// Check if the base layout exists - critical error if not
			if errors.Is(err, fs.ErrNotExist) {
				return nil, fmt.Errorf("base layout template 'html/layouts/base.tmpl' not found")
			}
			return nil, fmt.Errorf("error parsing layout for %s: %w", name, err)
		}

		// 5. Find and parse all partial templates (*.tmpl) into the set.
		partials, err := fs.Glob(ui.Files, "html/partials/*.tmpl") // CHANGED pattern
		if err != nil {
			// If Glob itself fails (other than not finding files)
			if !errors.Is(err, fs.ErrNotExist) && !strings.Contains(err.Error(), "no matching files found") {
				return nil, fmt.Errorf("error globbing partials for %s: %w", name, err)
			}
			// If no partials exist, err will be nil or ErrNotExist/match error, which is fine.
		}

		if len(partials) > 0 {
			// Parse all found partials into the same template set
			ts, err = ts.ParseFS(ui.Files, partials...)
			if err != nil {
				return nil, fmt.Errorf("error parsing partials for %s: %w", name, err)
			}
		}

		// 6. Add the fully parsed template set to the cache map.
		// The key is the page filename, e.g., "home.tmpl"
		cache[name] = ts
	}

	return cache, nil
}