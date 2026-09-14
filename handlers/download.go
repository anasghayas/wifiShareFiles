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

}
