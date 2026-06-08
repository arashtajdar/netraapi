package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sheedbox-api/config"
	"sheedbox-api/models"

	"github.com/go-chi/chi/v5"
)

func GetSeries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := `SELECT id, title, description, director, cast_members, rating, poster_url, backdrop_url, created_at, updated_at FROM series`
	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var seriesList []models.Series
	for rows.Next() {
		var s models.Series
		if err := rows.Scan(&s.ID, &s.Title, &s.Description, &s.Director, &s.CastMembers, &s.Rating, &s.PosterURL, &s.BackdropURL, &s.CreatedAt, &s.UpdatedAt); err == nil {
			seriesList = append(seriesList, s)
		}
	}
	if seriesList == nil {
		seriesList = []models.Series{}
	}
	json.NewEncoder(w).Encode(seriesList)
}

func GetSeriesDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, `{"error": "Missing series ID"}`, http.StatusBadRequest)
		return
	}

	var s models.Series
	query := `SELECT id, title, description, director, cast_members, rating, poster_url, backdrop_url, created_at, updated_at FROM series WHERE id = ?`
	err := config.DB.QueryRow(query, id).Scan(&s.ID, &s.Title, &s.Description, &s.Director, &s.CastMembers, &s.Rating, &s.PosterURL, &s.BackdropURL, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "Series not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}

	// Fetch Seasons
	seasonsQuery := `SELECT id, series_id, season_number, title, description, created_at FROM seasons WHERE series_id = ? ORDER BY season_number ASC`
	rows, err := config.DB.Query(seasonsQuery, s.ID)
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type SeasonDetail struct {
		models.Season
		Episodes []models.Episode `json:"episodes"`
	}

	seasonsList := []SeasonDetail{}
	for rows.Next() {
		var seas models.Season
		if err := rows.Scan(&seas.ID, &seas.SeriesID, &seas.SeasonNumber, &seas.Title, &seas.Description, &seas.CreatedAt); err == nil {
			seasonsList = append(seasonsList, SeasonDetail{Season: seas, Episodes: []models.Episode{}})
		}
	}

	// Fetch Episodes for each Season
	for i := range seasonsList {
		episodesQuery := `SELECT id, season_id, episode_number, title, description, video_sources, subtitles, intro_start, intro_end, created_at FROM episodes WHERE season_id = ? ORDER BY episode_number ASC`
		epRows, err := config.DB.Query(episodesQuery, seasonsList[i].ID)
		if err != nil {
			continue
		}
		defer epRows.Close()

		for epRows.Next() {
			var ep models.Episode
			if err := epRows.Scan(&ep.ID, &ep.SeasonID, &ep.EpisodeNumber, &ep.Title, &ep.Description, &ep.VideoSources, &ep.Subtitles, &ep.IntroStart, &ep.IntroEnd, &ep.CreatedAt); err == nil {
				seasonsList[i].Episodes = append(seasonsList[i].Episodes, ep)
			}
		}
	}

	response := struct {
		models.Series
		Seasons []SeasonDetail `json:"seasons"`
	}{
		Series:  s,
		Seasons: seasonsList,
	}

	json.NewEncoder(w).Encode(response)
}

