# 📁 wifiShare

> Share files easily with friends over campus WiFi — bypassing router client isolation using Go & Cloudflare Tunnels!

---

## 🔍 The Problem & How wifiShare Solves It

### Why direct IP connection / pinging fails on Campus WiFi:
On campus networks (like those managed by **Sophos**), **Client Isolation** is enabled. Even though you and your friend are on the same subnet with the same Default Gateway (`192.168.156.1`), the WiFi router blocks direct device-to-device communication. 

