// cmd/web/main.go
package main

import (
	"context"
	"database/sql"
	"errors" // Added for checking errors
	"flag"
	"html/template"
	"fmt"
	"log/slog"
	"net/http" // Required for http.Server
	"os"
	"time"

	_ "github.com/lib/pq"                                // PostgreSQL driver
	"github.com/mickali02/mood-notes-app/internal/data"  // Correct data package path
	_ "github.com/mickali02/mood-notes-app/ui"          // Import the ui package with embedded files
)

// application struct holds application-wide dependencies.
type application struct {
	logger        *slog.Logger
	addr          string
	moodNotes     *data.MoodNoteModel // Use the specific model
	templateCache map[string]*template.Template
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	// Read DSN from environment variable for better security/config management
	dsn := flag.String("dsn", os.Getenv("MOODNOTES_DB_DSN"), "PostgreSQL DSN (reads MOODNOTES_DB_DSN env var)")

	flag.Parse()

	// --- Logging ---
	// Use a structured logger with Debug level during development
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// --- Database ---
	// Check if the DSN was provided (either via flag or env var)
	if *dsn == "" {
		logger.Error("database DSN configuration is missing. Set the MOODNOTES_DB_DSN environment variable or use the -dsn flag.")
		os.Exit(1)
	}

	db, err := openDB(*dsn)
	if err != nil {
		logger.Error("database connection error", "error", err)
		os.Exit(1)
	}
	// Defer closing the connection pool when main() exits.
	defer db.Close()
	logger.Info("database connection pool established")

	// --- Template Cache ---
	// Initialize the template cache using the function from templates.go
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error("failed to create template cache", "error", err)
		os.Exit(1)
	}
	logger.Info("template cache loaded successfully")


	// --- Initialize Application Dependencies ---
	app := &application{
		logger:        logger,
		addr:          *addr,
		moodNotes:     &data.MoodNoteModel{DB: db}, // Initialize MoodNoteModel with the DB pool
		templateCache: templateCache,
	}

	// --- Start HTTP Server ---
	logger.Info("starting server", "address", app.addr)
	// Call the serve method defined in server.go
	err = app.serve()
	// Check for specific server closed error, which is expected on shutdown
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
	logger.Info("server stopped gracefully") // Log on graceful shutdown too
}

// openDB connects to the database and verifies the connection.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pool settings (important for performance and resource management)
	db.SetMaxOpenConns(25) // Max number of open connections
	db.SetMaxIdleConns(25) // Max number of connections sitting idle
	db.SetConnMaxIdleTime(5 * time.Minute) // How long a connection can be idle before being closed
	db.SetConnMaxLifetime(2 * time.Hour) // Max lifetime of any connection


	// Create a context with a timeout for the initial ping to verify the connection.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ping the database to verify the connection is live.
	err = db.PingContext(ctx)
	if err != nil {
		// If ping fails, close the pool before returning the error
		db.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return db, nil
}