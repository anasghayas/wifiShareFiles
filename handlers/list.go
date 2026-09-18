package handlers
import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)
// FileItem defines the JSON structure sent to the web browser
type FileItem struct {
	Name    string `json:"name"`
	Size    string `json:"size"`
	ModTime string `json:"modTime"`
}
// ListHandler scans the ./uploads directory and returns a JSON array of files
func ListHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Only allow HTTP GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	uploadDir := "./uploads"
	// 2. Ensure uploads directory exists
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		http.Error(w, "Could not access uploads directory", http.StatusInternalServerError)
		return
	}
	// 3. Read directory entries from disk
	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		http.Error(w, "Could not read uploads directory", http.StatusInternalServerError)
		return
	}
	// 4. Loop through entries and collect file metadata
	fileList := make([]FileItem, 0)
	for _, entry := range entries {
		// Ignore subdirectories — only list files
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fileList = append(fileList, FileItem{
			Name:    entry.Name(),
			Size:    formatFileSize(info.Size()),
			ModTime: info.ModTime().Format("2006-01-02 15:04"),
		})
	}
	// 5. Set response header to JSON format
	w.Header().Set("Content-Type", "application/json")
	// 6. Encode fileList slice into JSON and send to client
	json.NewEncoder(w).Encode(fileList)
}
// Helper function: Converts raw bytes (e.g. 2457600) into human readable string ("2.34 MB")
func formatFileSize(bytes int64) string {
	const (
		_  = iota
		KB = 1 << (10 * iota) // 1024 bytes
		MB                    // 1024 * 1024 bytes
		GB                    // 1024 * 1024 * 1024 bytes
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
