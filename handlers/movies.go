package handlers

import (
	"database/sql"
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
		var desc, relDate, dir, poster, backdrop sql.NullString
		var imdb, local sql.NullFloat64
		var cast, vid, sub []byte
		var introStart, introEnd sql.NullInt32
		
		if err := rows.Scan(
			&m.ID, &m.Title, &desc, &relDate, &dir,
			&cast, &imdb, &local, &poster,
			&backdrop, &vid, &sub, &introStart,
			&introEnd, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			continue
		}
		
		m.Description = desc.String
		if relDate.Valid { m.ReleaseDate = &relDate.String }
		if dir.Valid { m.Director = &dir.String }
		if cast != nil { m.CastMembers = cast } else { m.CastMembers = []byte("[]") }
		if imdb.Valid { m.IMDBRating = &imdb.Float64 }
		if local.Valid { m.LocalRating = &local.Float64 }
		if poster.Valid { m.PosterURL = &poster.String }
		if backdrop.Valid { m.BackdropURL = &backdrop.String }
		if vid != nil { m.VideoSources = vid } else { m.VideoSources = []byte("[]") }
		if sub != nil { m.Subtitles = sub } else { m.Subtitles = []byte("{}") }
		
		if introStart.Valid { 
			v := int(introStart.Int32)
			m.IntroStart = &v 
		}
		if introEnd.Valid { 
			v := int(introEnd.Int32)
			m.IntroEnd = &v 
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
