// cmd/web/render.go
package main

import (
	"bytes"
	"fmt"
	"net/http"
	"sync" // Added for sync.Pool
)

// bufferPool helps reuse buffers for template execution, improving performance.
var bufferPool = sync.Pool{
	New: func() any {
		// Provide a function to create a new buffer when the pool is empty.
		return new(bytes.Buffer) // More concise way to get *bytes.Buffer
	},
}

// renderTemplate retrieves a template from the cache, executes it with the provided data,
// and writes the result to the http.ResponseWriter.
// It handles potential errors during template lookup and execution.
func (app *application) renderTemplate(w http.ResponseWriter, status int, page string, data *TemplateData) error {
	// --- Template Lookup ---
	// Retrieve the requested template set ('page') from the application's template cache.
	ts, ok := app.templateCache[page]
	if !ok {
		// If the template doesn't exist in the cache, return an error.
		return fmt.Errorf("template %s does not exist", page)
	}

	// --- Buffer Management ---
	// Get a buffer from the pool. Type assertion is needed.
	buf := bufferPool.Get().(*bytes.Buffer)
	// Reset the buffer to ensure it's empty before use.
	buf.Reset()
	// Defer putting the buffer back into the pool. This runs *after* the function completes
	// or panics, ensuring the buffer is always returned for reuse.
	defer bufferPool.Put(buf)

	// --- Template Execution ---
	// Execute the template set. We execute the "base" template defined in the layout,
	// which will in turn include the specific page content.
	// Write the output into the buffer.
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		// Return the error if execution fails.
		return fmt.Errorf("failed to execute template %s: %w", page, err)
	}

	// --- Response Writing ---
	// If execution succeeded, set the HTTP status code header.
	// This *must* be done before writing to the response body.
	w.WriteHeader(status)

	// Write the contents of the buffer (the rendered HTML) to the http.ResponseWriter.
	_, err = buf.WriteTo(w)
	if err != nil {
		// Return error if writing to the response fails (e.g., connection closed).
		// Note: Header was already sent, so we can't change the status code here.
		return fmt.Errorf("failed to write template buffer to response: %w", err)
	}

	// Return nil on successful rendering and writing.
	return nil
}