package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"sheedbox-api/services"

	"github.com/go-chi/chi/v5"
)

// LiveTVHandler handles HTTP requests for the Live TV domain.
type LiveTVHandler struct {
	livetvService *services.LiveTVService
}

// NewLiveTVHandler creates a new LiveTVHandler.
func NewLiveTVHandler(livetvService *services.LiveTVService) *LiveTVHandler {
	return &LiveTVHandler{livetvService: livetvService}
}

func (h *LiveTVHandler) GetLiveChannels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	channels, err := h.livetvService.ListChannels(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(channels)
}

// YoutubeEmbed serves a YouTube player html with proper referrer policies to bypass Error 153.
func (h *LiveTVHandler) YoutubeEmbed(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing video ID", http.StatusBadRequest)
		return
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>YouTube Player</title>
    <style>
        body, html { margin: 0; padding: 0; width: 100%%; height: 100%%; overflow: hidden; background-color: #000; }
        iframe { width: 100%%; height: 100%%; border: none; }
    </style>
</head>
<body>
    <iframe 
        src="https://www.youtube.com/embed/%s?autoplay=1&mute=0&loop=1&playlist=%s&controls=1" 
        referrerpolicy="strict-origin-when-cross-origin"
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" 
        allowfullscreen>
    </iframe>
</body>
</html>`, id, id)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

