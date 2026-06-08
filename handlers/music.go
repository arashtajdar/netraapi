package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sheedbox-api/config"
	"sheedbox-api/models"

	"github.com/go-chi/chi/v5"
)

func GetMusic(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT id, title, description, artist, poster_url, backdrop_url, release_date FROM music_content ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var musicList []models.MusicContent
	for rows.Next() {
		var m models.MusicContent
		var releaseDate sql.NullString
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.Artist, &m.PosterURL, &m.BackdropURL, &releaseDate); err == nil {
			if releaseDate.Valid {
				m.ReleaseDate = &releaseDate.String
			}
			musicList = append(musicList, m)
		}
	}

	if musicList == nil {
		musicList = []models.MusicContent{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(musicList)
}

func GetMusicDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var m models.MusicContent
	var releaseDate sql.NullString

	err := config.DB.QueryRow(`
		SELECT id, title, description, artist, video_sources, audio_sources, poster_url, backdrop_url, release_date
		FROM music_content WHERE id = ?
	`, id).Scan(
		&m.ID, &m.Title, &m.Description, &m.Artist, &m.VideoSources, &m.AudioSources, &m.PosterURL, &m.BackdropURL, &releaseDate,
	)

	if err != nil {
		http.Error(w, "Music not found", http.StatusNotFound)
		return
	}

	if releaseDate.Valid {
		m.ReleaseDate = &releaseDate.String
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}
