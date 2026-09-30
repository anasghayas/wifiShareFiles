package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"shareFiles/handlers" // Imports our custom handlers package
)

func main() {
	// 0. Parse command-line flags
	//    WHAT: flag.String defines a command-line argument you can pass when starting the server.
	//    HOW:  go run main.go -password mysecret123
	//          flag.String("password", "", "...") means:
	//            - Flag name: "password" (used as -password on command line)
	//            - Default value: "" (empty = no auth, same as before)
	//            - Description: shown when you run go run main.go -help
	//    WHY:  We want the admin to choose their own password at startup,
	//          not hardcode it in the source code.
	password := flag.String("password", "", "Admin password for upload/delete (leave empty to disable auth)")
	flag.Parse() // Actually reads the command-line arguments

	// 1. Automatically create the 'uploads' directory on server startup
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}
	// WHAT: Register a "handler" for the root URL path "/"
	// WHY:  When someone visits http://localhost:8080/ in their browser,
	//       Go needs to know WHICH function to run. This line says:
	//       "Hey Go, when anyone visits '/', run this function."
	// HOW:  http.HandleFunc takes two things:
	//       1. A URL pattern ("/" means the homepage)
	//       2. A function to run when that URL is visited
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 'w' = ResponseWriter — we write our response INTO this (it goes to the browser)
		// 'r' = Request — contains everything about the incoming request (URL, method, headers, etc.)

		// fmt.Fprintf writes text into 'w', which sends it to the browser
		// Parse the HTML template from the 'templates' directory
		// Only serve index.html for exact path "/", otherwise return 404
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			// If template file is missing or corrupted, return a 500 Internal Server Error
			http.Error(w, "Could not load template: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Execute (render) the template into 'w' (the HTTP response to the browser)
		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Could not render template: "+err.Error(), http.StatusInternalServerError)
			return
		}
	})
	// 3. Register backend handlers from our 'handlers' package
	//    Upload and delete routes are WRAPPED with RequireAuth middleware.
	//    Download and file list are left OPEN — anyone can browse and download.
	//
	//    RequireAuth(handler, password) returns a NEW handler that:
	//      1. Checks X-Admin-Password header
	//      2. If correct → calls the original handler
	//      3. If wrong  → returns 401 Unauthorized
	//    If password is empty (""), RequireAuth lets everything through (no auth).

	// 🔒 Protected routes (require admin password)
	http.HandleFunc("/upload", handlers.RequireAuth(handlers.UploadHandler, *password))
	http.HandleFunc("/upload/chunk", handlers.RequireAuth(handlers.ChunkUploadHandler, *password))
	http.HandleFunc("/upload/complete", handlers.RequireAuth(handlers.ChunkCompleteHandler, *password))
	http.HandleFunc("/delete/", handlers.RequireAuth(handlers.DeleteHandler, *password))

	// 🔓 Open routes (no password needed — anyone can download/browse)
	http.HandleFunc("/download/", handlers.DownloadHandler)
	http.HandleFunc("/files", handlers.ListHandler)
	// WHAT: Start the HTTP server on port 8080
	// WHY:  The server needs to "listen" for incoming requests.
	//       Think of it like opening a shop — you set up inside, then open the door.
	//       ListenAndServe opens the door (port 8080) and waits for customers (requests).
	// HOW:  ":8080" means "listen on all network interfaces, port 8080"
	//       'nil' means "use the default router (the HandleFunc we set up above)"
	fmt.Println("🚀 wifiShare server starting...")
	fmt.Println("📁 Open your browser and go to: http://localhost:8080")
	fmt.Println("🛑 Press Ctrl+C to stop the server")

	// Print auth status so the admin knows if password protection is active
	if *password != "" {
		fmt.Println("🔐 Admin auth ENABLED — upload/delete require password")
	} else {
		fmt.Println("🔓 Admin auth DISABLED — anyone can upload/delete (use -password to enable)")
	}

	// 4. Start the background cleanup worker (deletes files older than 24 hours)
	//    'go' keyword = launch as a goroutine (lightweight background thread)
	//    This runs silently alongside the HTTP server — no extra terminal needed!
	go handlers.StartCleanupWorker()
	fmt.Println("🧹 Auto-cleanup worker started (deletes files older than 24 hours)")

	// log.Fatal: if the server fails to start (e.g., port already in use),
	// it prints the error and exits the program
	log.Fatal(http.ListenAndServe(":8080", nil))
}