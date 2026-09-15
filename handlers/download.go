package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DownloadHandler serves files stored in the ./uploads directory for download
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Only allow HTTP GET requests for downloading
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Extract filename from URL path (URL looks like: "/download/filename.ext")
	// strings.TrimPrefix removes "/download/" leaving just "filename.ext"
	filename := strings.TrimPrefix(r.URL.Path, "/download/")
	if filename == "" || filename == "/download/" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	// 3. Security: Sanitize filename to prevent Directory Traversal Attacks
	// WHAT: filepath.Base converts "../../secret.txt" -> "secret.txt"
	// WHY:  Prevents malicious users from accessing files outside the uploads folder!
	cleanFilename := filepath.Base(filename)
	filePath := filepath.Join("./uploads", cleanFilename)

	// 4. Check if the requested file actually exists on disk
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) || fileInfo.IsDir() {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// 5. Tell the browser to download the file instead of displaying it inline
	w.Header().Set("Content-Disposition", "attachment; filename=\""+cleanFilename+"\"")

	// 6. Efficiently stream the file to the browser
	// WHAT: http.ServeFile handles reading the file from disk and sending it over network
	// HOW:  It automatically handles content types, file streaming, and large files
	http.ServeFile(w, r, filePath)
}
