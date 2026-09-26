package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ChunkUploadHandler receives a single chunk of a large file upload
// ROUTE: POST /upload/chunk
//
// HOW IT WORKS:
// The browser splits a 3.2GB file into ~320 chunks of 10MB each.
// For each chunk, the browser sends a POST request with:
//   - "file"       : the 10MB blob of data
//   - "filename"   : the original full filename (e.g. "my_video.mp4")
//   - "chunkIndex" : which piece this is (0, 1, 2, ... 319)
//   - "totalChunks": how many total pieces there are (320)
//
// We save each chunk to: uploads/.chunks/{filename}/chunk_003
func ChunkUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parse the multipart form — allow up to 15MB per chunk in memory
	//    (our chunks are 10MB, so 15MB buffer gives some breathing room)
	const maxChunkMemory = 15 << 20 // 15MB
	if err := r.ParseMultipartForm(maxChunkMemory); err != nil {
		http.Error(w, "Failed to parse chunk: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Extract metadata from the form fields
	filename := r.FormValue("filename")       // Original filename like "video.mp4"
	chunkIndex := r.FormValue("chunkIndex")    // "0", "1", "2", etc.
	totalChunks := r.FormValue("totalChunks")  // Total number of chunks

	if filename == "" || chunkIndex == "" || totalChunks == "" {
		http.Error(w, "Missing required fields: filename, chunkIndex, totalChunks", http.StatusBadRequest)
		return
	}

	// 3. Security: sanitize the filename to prevent path traversal
	cleanFilename := filepath.Base(filename)

	// 4. Validate chunkIndex is a valid number
	chunkIdx, err := strconv.Atoi(chunkIndex)
	if err != nil || chunkIdx < 0 {
		http.Error(w, "Invalid chunkIndex", http.StatusBadRequest)
		return
	}

	// 5. Extract the chunk file data from the form
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read chunk data: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 6. Create the temporary chunk directory: uploads/.chunks/my_video.mp4/
	chunkDir := filepath.Join("./uploads", ".chunks", cleanFilename)
	if err := os.MkdirAll(chunkDir, os.ModePerm); err != nil {
		http.Error(w, "Failed to create chunk directory", http.StatusInternalServerError)
		return
	}

	// 7. Save this chunk as: uploads/.chunks/my_video.mp4/chunk_003
	//    fmt.Sprintf("%05d", 3) produces "00003" — zero-padded so files sort correctly
	chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%05d", chunkIdx))
	destFile, err := os.Create(chunkPath)
	if err != nil {
		http.Error(w, "Failed to save chunk: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	// 8. Stream the chunk data from the HTTP request to disk
	_, err = io.Copy(destFile, file)
	if err != nil {
		http.Error(w, "Failed to write chunk data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 9. Respond with success for this individual chunk
	fmt.Fprintf(w, "Chunk %s of %s received", chunkIndex, totalChunks)
}

// ChunkCompleteHandler merges all chunks into the final file
// ROUTE: POST /upload/complete
//
// HOW IT WORKS:
// After all 320 chunks have been uploaded, the browser sends one final request:
//   POST /upload/complete  with  filename=my_video.mp4
//
// This function:
// 1. Opens all chunk files in order (chunk_00000, chunk_00001, ...)
// 2. Creates the final destination file (uploads/my_video.mp4)
// 3. Copies each chunk's bytes into the final file sequentially
// 4. Deletes the temporary .chunks/my_video.mp4/ directory
func ChunkCompleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Parse form to get the filename
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	filename := r.FormValue("filename")
	if filename == "" {
		http.Error(w, "Missing filename", http.StatusBadRequest)
		return
	}

	cleanFilename := filepath.Base(filename)
	chunkDir := filepath.Join("./uploads", ".chunks", cleanFilename)

	// 2. Read all chunk files from the temp directory
	entries, err := os.ReadDir(chunkDir)
	if err != nil {
		http.Error(w, "Could not read chunks directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Filter for chunk files only and sort them in order
	var chunkFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "chunk_") {
			chunkFiles = append(chunkFiles, entry.Name())
		}
	}
	sort.Strings(chunkFiles) // Sorts: chunk_00000, chunk_00001, chunk_00002, ...

	if len(chunkFiles) == 0 {
		http.Error(w, "No chunks found for this file", http.StatusBadRequest)
		return
	}

	// 4. Create the final destination file
	destPath := filepath.Join("./uploads", cleanFilename)
	destFile, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "Failed to create final file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	// 5. Merge: open each chunk in order and copy its bytes into the final file
	for _, chunkName := range chunkFiles {
		chunkPath := filepath.Join(chunkDir, chunkName)

		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			http.Error(w, "Failed to open chunk: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// io.Copy streams bytes from chunk file → final file
		_, err = io.Copy(destFile, chunkFile)
		chunkFile.Close() // Close each chunk after copying (not deferred — we're in a loop)

		if err != nil {
			http.Error(w, "Failed to merge chunk: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// 6. Clean up: delete the entire temporary chunks directory
	os.RemoveAll(chunkDir)

	// 7. Respond with success!
	fmt.Fprintf(w, "File '%s' assembled successfully from %d chunks!", cleanFilename, len(chunkFiles))
}
