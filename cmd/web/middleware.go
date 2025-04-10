// cmd/web/middleware.go
package main

import (
	"net/http"
)

// loggingMiddleware logs details about incoming HTTP requests.
func (app *application) loggingMiddleware(next http.Handler) http.Handler {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract request details for logging.
		var (
			ip     = r.RemoteAddr
			proto  = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
		)
		// Log the extracted details using the application's structured logger.
		app.logger.Info("received request", "ip", ip, "protocol", proto, "method", method, "uri", uri)

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)

		// Optional: Log after the request is processed (e.g., include status code if using a response recorder middleware)
		app.logger.Info("request processed", "method", method, "uri", uri) // Simple processed message
	})
	return fn
}

// Add other middleware here later (e.g., recoverPanic, authenticate)
/*
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
*/