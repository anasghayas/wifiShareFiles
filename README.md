# 📁 wifiShare

> Share files easily with friends over campus WiFi — bypassing router client isolation using Go & Cloudflare Tunnels!

---

## 🔍 The Problem & How wifiShare Solves It

### Why direct IP connection / pinging fails on Campus WiFi:
On campus networks (like those managed by **Sophos**), **Client Isolation** is enabled. Even though you and your friend are on the same subnet with the same Default Gateway (`192.168.156.1`), the WiFi router blocks direct device-to-device communication. 

### The Solution:
`wifiShare` runs a lightweight **Go HTTP server** on your laptop and uses **Cloudflare Quick Tunnels** to create a secure public relay URL (`https://<random-id>.trycloudflare.com`). 
Your friend simply opens that URL in their browser to upload and download files — no installation needed on their end!

---

## ✨ Features

- 🎨 **Modern UI**: Dark theme glassmorphism design with drag-and-drop file upload.
- ⚡ **Go-Powered Backend**: Fast, lightweight HTTP server using Go standard library (`net/http`).
- 📦 **Chunked Uploads (up to 5GB)**: Files are split into 10MB chunks in the browser and reassembled on the server. Bypasses Cloudflare's 100MB request limit and survives flaky WiFi — each chunk auto-retries up to 3 times!
- 🗑️ **Instant File Delete**: One-click delete button next to each file removes it from disk immediately.
- ⏰ **24-Hour Auto-Cleanup**: A background goroutine silently scans `uploads/` every hour and deletes files older than 24 hours — no cron job needed.
- 📊 **Live Progress Bar**: Real-time upload progress showing chunk count (`"Uploading... 45% (145/320 chunks)"`).
- 🔒 **Secure**: Path traversal protection (`filepath.Base`) on all endpoints to prevent unauthorized file access.
- 🔐 **Admin Auth**: Password-protect uploads & deletes with a `-password` flag. Friends can still browse and download freely.
- 📱 **Mobile-Friendly**: Responsive design that works on phones — upload from mobile data or WiFi.

---

## 🚀 Quick Start

### Prerequisites
- [Go](https://go.dev/dl/) (1.21+)
- [Cloudflared](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/) (for sharing outside your WiFi)

### 1. Clone & Run the Server
```bash
git clone <your-repo-url>
cd wifiShare

# Without password (anyone can upload/delete):
go run main.go

# With password protection (recommended):
go run main.go -password yourSecretPassword
```

You should see:
```
🚀 wifiShare server starting...
📁 Open your browser and go to: http://localhost:8080
🛑 Press Ctrl+C to stop the server
🔐 Admin auth ENABLED — upload/delete require password
🧹 Auto-cleanup worker started (deletes files older than 24 hours)
```

### 2. Share with Friends (Cloudflare Tunnel)
In a **second terminal**, run:
```bash
cloudflared tunnel --url http://localhost:8080
```

Cloudflare gives you a public URL like:
```
https://some-random-words.trycloudflare.com
```

Send that URL to your friends — they can upload & download files from their browser!

---

## 📁 Project Structure

```text
wifiShare/
├── main.go                   # Entry point — server, routes & cleanup worker
├── handlers/
│   ├── auth.go               # RequireAuth middleware — password-checks protected routes
│   ├── upload.go             # POST /upload — single-request upload (small files)
│   ├── chunk_upload.go       # POST /upload/chunk & /upload/complete — chunked upload (large files)
│   ├── download.go           # GET /download/{filename} — file streaming
│   ├── delete.go             # DELETE /delete/{filename} — instant file removal
│   ├── list.go               # GET /files — JSON list of shared files
│   └── cleanup.go            # Background goroutine — auto-deletes files > 24 hours old
├── templates/
│   └── index.html            # Frontend UI (HTML / CSS / JS with chunked upload + admin bar)
├── uploads/                  # Shared files stored here (auto-created on startup)
│   └── .chunks/              # Temp directory for in-progress chunked uploads (auto-cleaned)
├── go.mod                    # Go module definition
└── README.md                 # You're reading this!
```

---

## 🔗 API Endpoints

| Method | Route | Handler | Auth | Description |
|--------|-------|---------|:----:|-------------|
| `GET` | `/` | inline in `main.go` | 🔓 | Serves the web UI |
| `POST` | `/upload` | `UploadHandler` | 🔒 | Single-request upload (files < 100MB) |
| `POST` | `/upload/chunk` | `ChunkUploadHandler` | 🔒 | Receives one 10MB chunk |
| `POST` | `/upload/complete` | `ChunkCompleteHandler` | 🔒 | Merges all chunks into final file |
| `GET` | `/download/{file}` | `DownloadHandler` | 🔓 | Streams a file for download |
| `DELETE` | `/delete/{file}` | `DeleteHandler` | 🔒 | Deletes a file from disk |
| `GET` | `/files` | `ListHandler` | 🔓 | Returns JSON array of uploaded files |

🔒 = Requires admin password &nbsp;&nbsp; 🔓 = Open to everyone

---

## 📤 How Chunked Uploads Work

```
Browser (JavaScript)                    Go Server
                                        
3.2GB Video File                        
  ├─ Chunk 0  (10MB) ──POST──►  Save to uploads/.chunks/video.mp4/chunk_00000
  ├─ Chunk 1  (10MB) ──POST──►  Save to uploads/.chunks/video.mp4/chunk_00001
  ├─ Chunk 2  (10MB) ──POST──►  Save to uploads/.chunks/video.mp4/chunk_00002
  ├─ ... (320 chunks total) ...
  └─ Chunk 319 (10MB) ──POST──► Save to uploads/.chunks/video.mp4/chunk_00319
                                        
"All done!" ──POST /upload/complete──►  Merge all 320 chunks → uploads/video.mp4
                                        Delete temp .chunks/ directory
                                        ──► ✅ "Upload Complete!"
```

**Why chunks?** Cloudflare's free tier blocks requests larger than ~100MB. By splitting into 10MB pieces, even a 5GB file uploads smoothly. Each chunk also auto-retries up to 3 times if campus WiFi drops a packet.

---

## 🧹 Auto-Cleanup (24-Hour File Deletion)

A background **goroutine** runs silently alongside the HTTP server:

```
main.go starts
  ├─► HTTP Server (foreground) — handles web requests
  └─► Cleanup Worker (background goroutine)
       └─ Every 1 hour: scan uploads/
          └─ If file age > 24 hours → delete it
```

**Important notes:**
- File age is tracked by **Windows/OS filesystem timestamps**, not the Go server. So if you stop the server for 12 hours, restart it, the worker will correctly see those files as 12+ hours old.
- The cleanup worker **only runs while `go run main.go` is running**. When you close the server, the janitor stops. But file ages are preserved — next startup catches up.
- Cloudflare is **not** involved in cleanup at all.

---

## ❓ FAQ

### Do I need Cloudflare running for everything?
**No!** Cloudflare is only needed if friends are outside your WiFi network. Everything else (uploads, downloads, deletes, auto-cleanup) works with just `go run main.go`.

| Feature | `go run main.go` | Cloudflare |
|---------|:-:|:-:|
| Upload / Download on same WiFi | ✅ Required | ❌ Not needed |
| Upload / Download from outside | ✅ Required | ✅ Required |
| 🗑️ Delete button | ✅ Required | ❌ Not needed |
| ⏰ 24hr auto-cleanup | ✅ Required | ❌ Not needed |
| 🔐 Admin auth | ✅ Required | ❌ Not needed |

### How does admin auth work?
Start the server with a password:
```bash
go run main.go -password yourSecretPassword
```
On the web page, type the password in the **Admin Bar** and click **Unlock**. Now you can upload and delete files. Without unlocking, uploads and deletes are blocked with a `401 Unauthorized` error. **Browsing and downloading are always open** — friends don't need the password to download files.

### Can friends still download if I set a password?
**Yes!** The password only protects **uploads** and **deletes**. The file list (`/files`) and downloads (`/download/`) are always open. Friends just can't upload new files or delete existing ones without the password.

### What if I forget to set a password?
If you run `go run main.go` without the `-password` flag, auth is **disabled** — everything works exactly like before, no password needed for anything. The terminal will show:
```
🔓 Admin auth DISABLED — anyone can upload/delete (use -password to enable)
```

### Can friends upload from mobile data?
**Yes!** As long as Cloudflare tunnel is running, anyone with the URL can upload — WiFi, mobile data, doesn't matter.

### Is the Cloudflare URL permanent?
**No.** Each time you run `cloudflared tunnel --url http://localhost:8080`, you get a new random URL. If you close that terminal, the URL dies. You need to share a fresh URL each session.

### What happens if my upload fails halfway?
The chunked upload system retries each 10MB chunk up to 3 times automatically. If it still fails, the partially uploaded chunks sit in `uploads/.chunks/` and will be cleaned up manually or on the next upload attempt of the same file.

---

## 🛠️ Built With
- **[Go](https://go.dev/)** — Backend server & file handling
- **[Cloudflare Quick Tunnels](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/do-more-with-tunnels/trycloudflare/)** — Public URL relay (free, no account needed)
- **Vanilla HTML/CSS/JS** — Frontend UI (no frameworks, no npm)

---

<p align="center">Made with ❤️ by <a href="https://github.com/anasghayas">anasghayas</a></p>