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