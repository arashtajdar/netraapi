package handlers

import (
	"encoding/json"
	"net/http"

	"sheedbox-api/services"
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
