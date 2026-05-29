package handlers

import (
	"encoding/json"
	"net/http"

	"netra-api/config"
	"netra-api/models"
)

func GetMovies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	category := r.URL.Query().Get("category")

	query := `SELECT id, title, description, video_url, thumbnail_url, category, duration_seconds FROM movies`
	var args []interface{}

	if category != "" {
		query += ` WHERE category = ?`
		args = append(args, category)
	}

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var m models.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.VideoURL, &m.ThumbnailURL, &m.Category, &m.DurationSeconds); err != nil {
			continue
		}
		movies = append(movies, m)
	}

	// Always return an empty array rather than null if empty
	if movies == nil {
		movies = []models.Movie{}
	}

	json.NewEncoder(w).Encode(movies)
}

func ResumePlayback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := r.Context().Value("user_id").(int)

	var input struct {
		MovieID               int `json:"movie_id"`
		ResumePositionSeconds int `json:"resume_position_seconds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Performs an UPSERT gracefully updating the position without crashing on duplicate key
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
