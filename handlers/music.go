package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"sheedbox-api/services"

	"github.com/go-chi/chi/v5"
)

type MusicHandler struct {
	musicService *services.MusicService
}

func NewMusicHandler(musicService *services.MusicService) *MusicHandler {
	return &MusicHandler{musicService: musicService}
}

func (h *MusicHandler) GetMusic(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	music, err := h.musicService.ListMusic(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(music)
}

func (h *MusicHandler) GetMusicDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, `{"error": "Missing music ID"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid music ID"}`, http.StatusBadRequest)
		return
	}

	m, err := h.musicService.GetMusicDetail(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}
	if m == nil {
		http.Error(w, `{"error": "Music not found"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(m)
}
