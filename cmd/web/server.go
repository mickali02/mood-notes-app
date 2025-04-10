// cmd/web/server.go
package main

import (
	"log/slog"
	"net/http"
	"time"
)

// serve configures and starts the application's HTTP server.
func (app *application) serve() error {
	// Configure the HTTP server.
	srv := &http.Server{
		Addr:    app.addr,            // Listen address from command-line flag/default
		Handler: app.routes(),        // Use the router returned by app.routes()
		ErrorLog: slog.NewLogLogger(app.logger.Handler(), slog.LevelError), // Use structured logger for server errors
		// Set timeouts to improve security and resource management.
		IdleTimeout:  time.Minute,      // Max time for idle connections
		ReadTimeout:  5 * time.Second,  // Max time to read request headers/body
		WriteTimeout: 10 * time.Second, // Max time to write response
	}

	// Log the server start address (this was already in main, but doesn't hurt here too for clarity).
	// app.logger.Info("starting server", "addr", srv.Addr) // This line is usually in main() before calling serve()

	// Start the HTTP server. ListenAndServe blocks until an error occurs
	// (e.g., port already in use) or the server is shut down gracefully.
	return srv.ListenAndServe()
}