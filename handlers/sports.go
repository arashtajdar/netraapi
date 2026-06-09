package handlers

import (
	"encoding/json"
	"net/http"

	"sheedbox-api/services"
)

// SportsHandler handles HTTP requests for the Sports domain.
type SportsHandler struct {
	sportsService *services.SportsService
}

// NewSportsHandler creates a new SportsHandler.
func NewSportsHandler(sportsService *services.SportsService) *SportsHandler {
	return &SportsHandler{sportsService: sportsService}
}

func (h *SportsHandler) GetSportsEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	events, err := h.sportsService.ListEvents(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(events)
}
