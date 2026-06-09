package handlers

import (
	"encoding/json"
	"net/http"

	"sheedbox-api/contextkeys"
	"sheedbox-api/services"
)

type WatchlistHandler struct {
	watchlistService *services.WatchlistService
}

func NewWatchlistHandler(watchlistService *services.WatchlistService) *WatchlistHandler {
	return &WatchlistHandler{watchlistService: watchlistService}
}

func (h *WatchlistHandler) GetUserWatchlists(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := contextkeys.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	lists, err := h.watchlistService.GetWatchlists(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(lists)
}
