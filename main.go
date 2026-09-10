package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
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
		fmt.Fprintf(w, "Hello from wifiShare! 🚀\nYour file sharing server is running.")
	})

	// WHAT: Start the HTTP server on port 8080
	// WHY:  The server needs to "listen" for incoming requests.
	//       Think of it like opening a shop — you set up inside, then open the door.
	//       ListenAndServe opens the door (port 8080) and waits for customers (requests).
	// HOW:  ":8080" means "listen on all network interfaces, port 8080"
	//       'nil' means "use the default router (the HandleFunc we set up above)"
	fmt.Println("🚀 wifiShare server starting...")
	fmt.Println("📁 Open your browser and go to: http://localhost:8080")
	fmt.Println("🛑 Press Ctrl+C to stop the server")

	// log.Fatal: if the server fails to start (e.g., port already in use),
	// it prints the error and exits the program
	log.Fatal(http.ListenAndServe(":8080", nil))
}