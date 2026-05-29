package handlers

import (
	"encoding/json"
	"net/http"

	"netra-api/config"
	"netra-api/models"
)

func GetMovies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := `SELECT id, title, description, release_date, director, cast_members, imdb_rating, local_rating, poster_url, backdrop_url, video_sources, subtitles, intro_start, intro_end, created_at, updated_at FROM movies`
	
	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var m models.Movie
		if err := rows.Scan(
			&m.ID, &m.Title, &m.Description, &m.ReleaseDate, &m.Director, 
			&m.CastMembers, &m.IMDBRating, &m.LocalRating, &m.PosterURL, 
			&m.BackdropURL, &m.VideoSources, &m.Subtitles, &m.IntroStart, 
			&m.IntroEnd, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			continue
		}
		movies = append(movies, m)
	}

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
