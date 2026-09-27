package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DeleteHandler removes a file from the uploads directory
// ROUTE: DELETE /delete/{filename}
//
// HOW IT WORKS:
// 1. Browser sends: DELETE /delete/my_video.mp4
// 2. We extract "my_video.mp4" from the URL
// 3. We sanitize the filename (security — prevent accessing files outside uploads/)
// 4. We delete the file from disk using os.Remove()
// 5. We respond with success or error
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Only allow DELETE method
	//    WHY: Using the correct HTTP method makes the API clear and predictable.
	//    GET = read, POST = create, DELETE = remove. This is called "RESTful" design.
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Extract filename from URL path
	//    URL: /delete/my_video.mp4 → TrimPrefix → "my_video.mp4"
	filename := strings.TrimPrefix(r.URL.Path, "/delete/")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	// 3. Security: sanitize filename to prevent path traversal attacks
	//    "../../../etc/passwd" → "passwd" (filepath.Base strips directory paths)
	cleanFilename := filepath.Base(filename)
	filePath := filepath.Join("./uploads", cleanFilename)

	// 4. Check if the file actually exists before trying to delete
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	// Don't allow deleting directories (only files)
	if info.IsDir() {
		http.Error(w, "Cannot delete directories", http.StatusBadRequest)
		return
	}

	// 5. Delete the file from disk
	//    os.Remove() permanently deletes the file — it does NOT go to Recycle Bin!
	err = os.Remove(filePath)
	if err != nil {
		http.Error(w, "Failed to delete file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 6. Respond with success
	fmt.Fprintf(w, "File '%s' deleted successfully!", cleanFilename)
}
