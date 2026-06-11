package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"sheedbox-api/config"
	"sheedbox-api/contextkeys"
	"sheedbox-api/services"

	"github.com/go-chi/chi/v5"
)

// MovieHandler handles HTTP requests for the Movie domain.
type MovieHandler struct {
	movieService *services.MovieService
}

// NewMovieHandler creates a new MovieHandler.
func NewMovieHandler(movieService *services.MovieService) *MovieHandler {
	return &MovieHandler{movieService: movieService}
}

// GetMovies returns the list of all movies.
func (h *MovieHandler) GetMovies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	movies, err := h.movieService.ListMovies(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(movies)
}

// GetMovieDetail returns the details of a specific movie.
func (h *MovieHandler) GetMovieDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, `{"error": "Missing movie ID"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid movie ID"}`, http.StatusBadRequest)
		return
	}

	m, err := h.movieService.GetMovieDetail(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}
	if m == nil {
		http.Error(w, `{"error": "Movie not found"}`, http.StatusNotFound)
		return
	}

	userLevel := contextkeys.UserLevelFromContext(r.Context())
	if m.AccessLevel > userLevel {
		http.Error(w, `{"error": "You don't have access to this content due to user level restrictions"}`, http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(m)
}

// ResumePlayback handles saving the user's current watch position.
// TODO: Move this to a dedicated WatchHistoryService/Repository in the future.
func (h *MovieHandler) ResumePlayback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := contextkeys.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var input struct {
		MovieID               int `json:"movie_id"`
		ResumePositionSeconds int `json:"resume_position_seconds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO user_watch_history (user_id, movie_id, resume_position_seconds) 
		VALUES (?, ?, ?) 
		ON DUPLICATE KEY UPDATE resume_position_seconds = ?, updated_at = CURRENT_TIMESTAMP
	`
	_, err := config.DB.Exec(query, userID, input.MovieID, input.ResumePositionSeconds, input.ResumePositionSeconds)
	if err != nil {
		http.Error(w, `{"error": "Failed to save position"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Progress synchronized gracefully"})
}

