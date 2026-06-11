package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"sheedbox-api/contextkeys"
	"sheedbox-api/services"

	"github.com/go-chi/chi/v5"
)

// SeriesHandler handles HTTP requests for the Series domain.
type SeriesHandler struct {
	seriesService *services.SeriesService
}

// NewSeriesHandler creates a new SeriesHandler.
func NewSeriesHandler(seriesService *services.SeriesService) *SeriesHandler {
	return &SeriesHandler{seriesService: seriesService}
}

func (h *SeriesHandler) GetSeries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	seriesList, err := h.seriesService.ListSeries(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(seriesList)
}

func (h *SeriesHandler) GetSeriesDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, `{"error": "Missing series ID"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid series ID"}`, http.StatusBadRequest)
		return
	}

	detail, err := h.seriesService.GetSeriesDetail(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}
	if detail == nil {
		http.Error(w, `{"error": "Series not found"}`, http.StatusNotFound)
		return
	}

	userLevel := contextkeys.UserLevelFromContext(r.Context())
	if detail.AccessLevel > userLevel {
		http.Error(w, `{"error": "You don't have access to this content due to user level restrictions"}`, http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(detail)
}

