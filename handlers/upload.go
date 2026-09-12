package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// UploadHandler processes POST requests for file uploads
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Only allow HTTP POST method for file uploads
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Set max upload size to 50MB
	// WHAT: 50 << 20 means 50 shifted left by 20 bits = 50 * 1024 * 1024 = 52,428,800 bytes (50 MB)
	// WHY:  Protects your server from running out of RAM if someone uploads a huge file.
	const maxUploadSize = 50 << 20 // 50MB
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "File too large. Max limit is 50MB", http.StatusBadRequest)
		return
	}

	// 3. Extract the file from the request form under key "file"
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file upload request: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close() // Close the file stream when function ends

	// 4. Ensure the destination "uploads" directory exists
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		http.Error(w, "Failed to create uploads directory", http.StatusInternalServerError)
		return
	}

	// 5. Clean filename to prevent path traversal attacks (e.g. "../../etc/passwd")
	filename := filepath.Base(header.Filename)
	destPath := filepath.Join(uploadDir, filename)

	// 6. Create the empty destination file on your hard disk
	destFile, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "Could not create destination file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	// 7. Stream content from uploaded network request to disk file
	_, err = io.Copy(destFile, file)
	if err != nil {
		http.Error(w, "Failed to save file content: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 8. Respond to the browser with success
	fmt.Fprintf(w, "File '%s' uploaded successfully!", filename)
}
