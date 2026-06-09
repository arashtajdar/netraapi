package handlers

import (
	"encoding/json"
	"net/http"
	"sheedbox-api/services"
	"sheedbox-api/services/storage"
)

func UploadMedia(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(50 << 20) // 50MB limit per request
	if err != nil {
		http.Error(w, "File too large or invalid multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file from 'file' field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if storage.ActiveProvider == nil {
		http.Error(w, "Storage provider not configured in backend", http.StatusInternalServerError)
		return
	}

	url, err := storage.ActiveProvider.UploadFile(file, header)
	if err != nil {
		http.Error(w, "Upload failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Queue for FFmpeg processing
	services.QueueVideoForProcessing(services.VideoTask{
		VideoURL: url,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": url})
}
