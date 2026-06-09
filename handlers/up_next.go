package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"sheedbox-api/models"
)

type UpNextHandler struct {
	db *sql.DB
}

func NewUpNextHandler(db *sql.DB) *UpNextHandler {
	return &UpNextHandler{db: db}
}

func (h *UpNextHandler) GetUpNext(w http.ResponseWriter, r *http.Request) {
	contentType := r.URL.Query().Get("type")
	id := r.URL.Query().Get("id")

	if contentType == "" || id == "" {
		http.Error(w, "type and id are required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// 1. Get up_next_timer setting
	var timerSetting string
	err := h.db.QueryRowContext(r.Context(), "SELECT setting_value FROM app_settings WHERE setting_key = 'up_next_timer'").Scan(&timerSetting)
	if err != nil {
		timerSetting = "10" // default
	}

	timerSetting = trimQuotes(timerSetting)

	response := map[string]interface{}{
		"timer": timerSetting,
	}

	// 2. Logic based on content type
	if contentType == "series" {
		// Find next episode
		var nextEp models.Episode
		// Simplified query: getting next episode in same season or season+1
		err := h.db.QueryRowContext(r.Context(), `
			SELECT e.id, e.season_id, e.episode_number, e.title, e.description, e.video_sources, e.subtitles, e.intro_start, e.intro_end, e.created_at
			FROM episodes e
			WHERE e.id > ? 
			ORDER BY e.id ASC LIMIT 1`, id).Scan(&nextEp.ID, &nextEp.SeasonID, &nextEp.EpisodeNumber, &nextEp.Title, &nextEp.Description, &nextEp.VideoSources, &nextEp.Subtitles, &nextEp.IntroStart, &nextEp.IntroEnd, &nextEp.CreatedAt)

		if err == nil {
			response["type"] = "episode"
			response["content"] = nextEp
			json.NewEncoder(w).Encode(response)
			return
		}
	} else if contentType == "movie" {
		// Recommend random movie
		var nextMovie models.Movie
		err := h.db.QueryRowContext(r.Context(), `
			SELECT id, title, description, poster_url, backdrop_url 
			FROM movies 
			WHERE id != ? 
			ORDER BY RAND() LIMIT 1`, id).Scan(&nextMovie.ID, &nextMovie.Title, &nextMovie.Description, &nextMovie.PosterURL, &nextMovie.BackdropURL)

		if err == nil {
			response["type"] = "movie"
			response["content"] = nextMovie
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Fallback empty response
	json.NewEncoder(w).Encode(response)
}

func trimQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
