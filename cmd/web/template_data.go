// cmd/web/template_data.go
package main

import (
	"time" // Added for CurrentYear

	"github.com/mickali02/mood-notes-app/internal/data"
	"github.com/mickali02/mood-notes-app/internal/validator"
)

// TemplateData holds data passed to HTML templates.
type TemplateData struct {
	CurrentYear int              // Example: To display in footer
	Flash       string           // For success/error messages (implement later with sessions)
	Notes       []*data.MoodNote // For the home page list
	Note        *data.MoodNote   // For pre-filling the edit form

	// Form Handling - use 'any' for flexibility or specific structs
	// This allows passing either MoodNoteCreateForm or MoodNoteEditForm
	Form any

	// You might add other general page data here later, e.g.,
	// IsAuthenticated bool
}

// newTemplateData creates a default TemplateData object.
func newTemplateData() *TemplateData {
	return &TemplateData{
		CurrentYear: time.Now().Year(), // Example default value
	}
}

// --- Form Structs (for type safety and clarity in handlers/templates) ---

// MoodNoteCreateForm holds the data submitted from the new note form + validation.
type MoodNoteCreateForm struct {
	Title   string `form:"title"`   // Tag matches form field name
	Content string `form:"content"` // Tag matches form field name
	// Embed validator to carry validation errors.
	validator.Validator
}

// MoodNoteEditForm holds data for editing, including ID and Version + validation.
type MoodNoteEditForm struct {
	ID      int64 `form:"id"`      // From hidden form field or URL param
	Title   string `form:"title"`
	Content string `form:"content"`
	Version int    `form:"version"` // From hidden form field for optimistic locking
	// Embed validator to carry validation errors.
	validator.Validator
}