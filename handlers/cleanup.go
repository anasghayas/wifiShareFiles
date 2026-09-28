package handlers

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StartCleanupWorker runs in the background (as a goroutine) and deletes
// files from the uploads/ directory that are older than 24 hours.
//
// HOW IT WORKS:
//   main.go calls:  go handlers.StartCleanupWorker()
//
//   The 'go' keyword launches this function in a GOROUTINE — a lightweight
//   background thread that runs alongside the HTTP server. Think of it like
//   hiring a janitor who silently checks the uploads folder every hour
//   and throws away anything older than 24 hours.
//
//   ┌─────────────────────────────────────────────────┐
//   │  main.go starts                                 │
//   │    ├─► HTTP Server (foreground) — serves pages  │
//   │    └─► Cleanup Worker (background goroutine)    │
//   │         └─ Every 1 hour: scan uploads/          │
//   │            └─ If file age > 24hr → delete it    │
//   └─────────────────────────────────────────────────┘
//
// WHY A GOROUTINE?
//   Go's goroutines are incredibly lightweight (only ~2KB of memory each).
//   Unlike threads in Java/C++, you can launch thousands of goroutines
//   without performance issues. For our use case, we just need one
//   background worker running alongside the server.
func StartCleanupWorker() {
	// time.NewTicker creates a channel that sends a "tick" every hour.
	// Think of it like an alarm clock that rings every 60 minutes.
	ticker := time.NewTicker(1 * time.Hour)

	// This loop runs FOREVER (until the server is stopped with Ctrl+C)
	for {
		// Wait for the next tick (blocks here for 1 hour)
		<-ticker.C

		log.Println("🧹 Cleanup worker: scanning uploads/ for files older than 24 hours...")
		cleanOldFiles("./uploads", 24*time.Hour)
	}
}

// cleanOldFiles scans a directory and deletes any file older than maxAge.
//
// HOW IT WORKS:
//   1. Read all entries in the directory
//   2. For each file (skip directories), check its ModTime (last modified time)
//   3. If time.Since(modTime) > 24 hours → delete it
//
// WHAT IS time.Since()?
//   time.Since(t) returns the duration elapsed since time 't'.
//   Example: if a file was last modified yesterday at 3pm, and it's now 4pm today,
//   time.Since(modTime) = 25 hours, which is > 24 hours → delete!
func cleanOldFiles(dir string, maxAge time.Duration) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("🧹 Cleanup: could not read directory %s: %v", dir, err)
		return
	}

	deletedCount := 0

	for _, entry := range entries {
		// Skip directories (like .chunks/) — we only delete regular files
		if entry.IsDir() {
			continue
		}

		// Skip hidden files (starting with '.')
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Get file metadata (includes modification time)
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Check if the file is older than maxAge (24 hours)
		fileAge := time.Since(info.ModTime())
		if fileAge > maxAge {
			filePath := filepath.Join(dir, entry.Name())
			err := os.Remove(filePath)
			if err != nil {
				log.Printf("🧹 Cleanup: failed to delete %s: %v", entry.Name(), err)
			} else {
				log.Printf("🧹 Cleanup: deleted '%s' (age: %s)", entry.Name(), fileAge.Round(time.Minute))
				deletedCount++
			}
		}
	}

	if deletedCount > 0 {
		log.Printf("🧹 Cleanup: removed %d file(s) older than 24 hours", deletedCount)
	} else {
		log.Println("🧹 Cleanup: no old files found — nothing to delete")
	}
}
