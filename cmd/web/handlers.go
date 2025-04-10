// cmd/web/handlers.go
package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/mickali02/mood-notes-app/internal/data"
	"github.com/mickali02/mood-notes-app/internal/validator"
)

// --- Helper Functions ---

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	method := r.Method
	uri := r.URL.RequestURI()
	app.logger.Error("internal server error", "method", method, "uri", uri, "error", err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, td *TemplateData) {
	if td == nil {
		td = newTemplateData()
	}
	// td.Flash = app.sessionManager.PopString(r.Context(), "flash") // Add later
	err := app.renderTemplate(w, status, page, td)
	if err != nil {
		app.logger.Error("error rendering template", "template", page, "error", err)
		app.serverError(w, r, err)
	}
}

// --- Route Handlers ---

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Call the correct model method
	notes, err := app.moodNotes.GetAll()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data := newTemplateData()
	data.Notes = notes
	app.render(w, r, http.StatusOK, "home.tmpl", data)
}

func (app *application) showMoodNoteForm(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	td := newTemplateData()

	if idStr == "" { // CREATE
		td.Form = MoodNoteCreateForm{}
		app.render(w, r, http.StatusOK, "note_form.tmpl", td)
	} else { // EDIT
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id < 1 {
			app.notFound(w) // Invalid ID format
			return
		}
		// Call the correct model method
		note, err := app.moodNotes.Get(id)
		if err != nil {
			// ** CORRECTED ERROR CHECK **
			// Check if the error message matches the one returned by mood_notes.go Get method
			if err.Error() == "mood note record not found" || err.Error() == "invalid mood note ID provided" {
				app.notFound(w)
			} else {
				app.serverError(w, r, err) // Handle other unexpected errors
			}
			return
		}
		// Populate form for editing
		td.Form = MoodNoteEditForm{
			ID:      note.ID,
			Title:   note.Title,
			Content: note.Content,
			Version: note.Version,
		}
		td.Note = note // Pass the full note data too
		app.render(w, r, http.StatusOK, "note_form.tmpl", td)
	}
}

func (app *application) createMoodNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form := MoodNoteCreateForm{
		Title:     r.PostForm.Get("title"),
		Content:   r.PostForm.Get("content"),
		Validator: *validator.NewValidator(),
	}
	noteToValidate := &data.MoodNote{Title: form.Title, Content: form.Content}
	// Use the standalone validation function from the data package
	data.ValidateMoodNote(&form.Validator, noteToValidate)

	if !form.ValidData() {
		td := newTemplateData()
		td.Form = form // Pass form with errors back
		app.render(w, r, http.StatusUnprocessableEntity, "note_form.tmpl", td)
		return
	}
	// Call the correct model method
	noteToInsert := &data.MoodNote{Title: form.Title, Content: form.Content}
	err = app.moodNotes.Insert(noteToInsert)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) updateMoodNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	err = r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	version, err := strconv.Atoi(r.PostForm.Get("version"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form := MoodNoteEditForm{
		ID:        id,
		Title:     r.PostForm.Get("title"),
		Content:   r.PostForm.Get("content"),
		Version:   version,
		Validator: *validator.NewValidator(),
	}
	noteToValidate := &data.MoodNote{ID: form.ID, Title: form.Title, Content: form.Content, Version: form.Version}
	// Use the standalone validation function
	data.ValidateMoodNote(&form.Validator, noteToValidate)

	if !form.ValidData() {
		td := newTemplateData()
		td.Form = form // Pass form with errors back
		app.render(w, r, http.StatusUnprocessableEntity, "note_form.tmpl", td)
		return
	}
	// Call the correct model method
	noteToUpdate := &data.MoodNote{ID: form.ID, Title: form.Title, Content: form.Content, Version: form.Version}
	err = app.moodNotes.Update(noteToUpdate)
	if err != nil {
		// ** CORRECTED ERROR CHECK **
		// Check the specific error messages returned by the model's Update method
		errMsg := err.Error()
		if errMsg == "mood note record not found or version mismatch" {
			// This could be not found OR edit conflict based on the model logic
			// Try to get the latest version to show the conflict message
			latestNote, getErr := app.moodNotes.Get(id) // Use Get again
			if getErr != nil && getErr.Error() == "mood note record not found" {
				// If Get also says not found, then it really wasn't there
				app.notFound(w)
			} else if getErr != nil {
				// Error fetching during conflict check
				app.serverError(w, r, fmt.Errorf("update failed (not found/conflict) and could not refetch note %d: %w", id, getErr))
			} else {
				// Successfully fetched latestNote, it was an edit conflict
				td := newTemplateData()
				form.Version = latestNote.Version // Update form version
				form.AddError("_conflict", "Edit Conflict: This note was updated by someone else. Please review the changes and try submitting again.")
				td.Form = form
				td.Note = latestNote
				app.render(w, r, http.StatusConflict, "note_form.tmpl", td) // 409 Conflict
			}
		} else {
			// Handle other unexpected errors from Update
			app.serverError(w, r, err)
		}
		return // Stop processing on error
	}
	// Success
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) deleteMoodNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	// Call the correct model method
	err = app.moodNotes.Delete(id)
	if err != nil {
		// ** CORRECTED ERROR CHECK **
		// Check the specific error message returned by the model's Delete method
		if err.Error() == "mood note record not found or already deleted" || err.Error() == "invalid mood note ID provided" {
			app.notFound(w)
		} else {
			// Handle other unexpected errors
			app.serverError(w, r, err)
		}
		return
	}
	// Success
	http.Redirect(w, r, "/", http.StatusSeeOther)
}