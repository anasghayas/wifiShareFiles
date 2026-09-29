package handlers

import (
	"net/http"
)

// RequireAuth is a MIDDLEWARE — a function that wraps another handler function.
//
// WHAT IS MIDDLEWARE?
//   Normally, a request goes straight to the handler:
//     Request ──────────────────► UploadHandler ──► Save file
//
//   With middleware, it passes through a "gatekeeper" first:
//     Request ──► RequireAuth() ──► UploadHandler ──► Save file
//                      │
//                      ├─ Password correct? → ✅ Let it through
//                      └─ Password wrong?   → ❌ Block with 401
//
// HOW IT WORKS IN GO:
//   This function takes a handler function as input and RETURNS a new handler function.
//   The returned function checks the password first, then calls the original handler.
//   This pattern is called a "higher-order function" — a function that takes/returns functions.
//
// USAGE IN main.go:
//   Instead of:  http.HandleFunc("/upload", handlers.UploadHandler)
//   We write:    http.HandleFunc("/upload", handlers.RequireAuth(handlers.UploadHandler, password))
//
// PARAMETERS:
//   next     — the original handler to call IF the password is correct
//   password — the admin password set via command-line flag
func RequireAuth(next http.HandlerFunc, password string) http.HandlerFunc {
	// This RETURNS a new function — the "wrapped" version of 'next'
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. If no password was set when starting the server, skip auth entirely.
		//    This means: go run main.go (without -password flag) = NO protection
		//    This keeps backward compatibility — server works the same as before
		//    unless you explicitly set a password.
		if password == "" {
			next(w, r) // No password configured, let everything through
			return
		}

		// 2. Read the password from the request header
		//    The browser (JavaScript) sends this header with every upload/delete request:
		//      X-Admin-Password: mysecret123
		//
		//    WHAT IS A HEADER?
		//    HTTP headers are key-value pairs sent alongside every request.
		//    Common headers: "Content-Type", "User-Agent", "Authorization".
		//    We use a custom header "X-Admin-Password" (the X- prefix means "custom/non-standard").
		providedPassword := r.Header.Get("X-Admin-Password")

		// 3. Compare the provided password with the server's password
		//    If they don't match (or no password was sent), block the request.
		if providedPassword != password {
			// Set CORS header so the browser's JavaScript can read the 401 status
			// (without this, fetch() might not see the error properly)
			w.Header().Set("Content-Type", "text/plain")

			// 401 Unauthorized = "You need to prove who you are"
			// (vs 403 Forbidden = "I know who you are but you're not allowed")
			http.Error(w, "Wrong or missing admin password", http.StatusUnauthorized)
			return
		}

		// 4. Password is correct! Call the original handler.
		next(w, r)
	}
}
