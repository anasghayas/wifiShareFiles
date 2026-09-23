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
- 🔒 **Secure**: Path traversal protection (`filepath.Base`) to prevent unauthorized file access.
- 📊 **Progress Bar**: Real-time upload progress tracking.
- 📦 **File Management**: Auto-formatting file sizes (KB, MB, GB) and file icons.
- 🛑 **50MB Upload Limit**: Enforced upload memory limit per request.
---