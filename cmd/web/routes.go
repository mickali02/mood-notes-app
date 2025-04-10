// cmd/web/routes.go
package main

import (
	"net/http"
	"os" 

	"github.com/mickali02/mood-notes-app/ui" // Import the ui package with embedded files
)

// neuterFileSystem prevents directory listing by returning an error for directories.
type neuterFileSystem struct {
	fs http.FileSystem
}

func (nfs neuterFileSystem) Open(path string) (http.File, error) {
	// Open the file from the underlying (embedded) filesystem.
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err // Pass through errors like file not found.
	}

	// Check if the opened item is a directory.
	s, err := f.Stat()
	if err != nil {
		// Error getting stats, close the file and return the error.
		f.Close()
		return nil, err
	}
	if s.IsDir() {
		// It's a directory, close it and return an error that indicates
		// it doesn't exist as a servable file (prevents directory listing).
		f.Close()
		return nil, os.ErrNotExist
	}

	// It's a file, return it.
	return f, nil
}

// routes defines and returns the application's HTTP request multiplexer.
func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// --- Static Files ---
	// Create a filesystem handler rooted at the embedded ui.Files.
	// Wrap it with neuterFileSystem to prevent directory listings.
	// http.FS converts embed.FS to the required http.FileSystem interface.
	embeddedFS := http.FS(ui.Files)
	fileServer := http.FileServer(neuterFileSystem{embeddedFS})

	// Handle requests for "/static/". Since our embed directive includes paths like
	// 'static/css/styles.css', requests for '/static/css/styles.css' will correctly
	// look for 'static/css/styles.css' within the embedded filesystem root.
	// No StripPrefix is needed here because the request path already matches the
	// structure within the embedded FS.
	mux.Handle("GET /static/", fileServer)

	// --- Mood Note Dynamic Routes ---
	mux.HandleFunc("GET /{$}", app.home)                   // Home page (list notes)
	mux.HandleFunc("GET /note/new", app.showMoodNoteForm)  // Show form to CREATE note
	mux.HandleFunc("POST /note/new", app.createMoodNote) // Handle form submission for CREATE

	// Use Go 1.22+ path parameters {id}
	mux.HandleFunc("GET /note/edit/{id}", app.showMoodNoteForm) // Show form to EDIT note
	mux.HandleFunc("POST /note/edit/{id}", app.updateMoodNote) // Handle form submission for UPDATE
	mux.HandleFunc("POST /note/delete/{id}", app.deleteMoodNote) // Handle deletion

	// --- Middleware ---
	// Apply middleware. Logging middleware is applied last (runs first).
	// Add other middleware like recovery, authentication later inside loggingMiddleware.
	return app.loggingMiddleware(mux)
}