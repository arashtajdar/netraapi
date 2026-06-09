package handlers

import (
	"encoding/json"
	"net/http"
	"sheedbox-api/config"
	"sheedbox-api/models"
	"sync"
	"time"
	"database/sql"
)

type CachedRecommendation struct {
	Movies    []models.Movie
	ExpiresAt time.Time
}

var recommendationCache = sync.Map{}

func GetRecommendations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	profileID := 0
	if pid := r.Context().Value("profile_id"); pid != nil {
		profileID = pid.(int)
	}
	userID := r.Context().Value("user_id").(int)
	
	cacheKey := userID
	if profileID != 0 {
		cacheKey = profileID
	}

	if cached, ok := recommendationCache.Load(cacheKey); ok {
		c := cached.(CachedRecommendation)
		if time.Now().Before(c.ExpiresAt) {
			json.NewEncoder(w).Encode(c.Movies)
			return
		}
	}

	var rows *sql.Rows
	var err error

	if profileID != 0 {
		query := `
			SELECT DISTINCT m.id, m.title, m.description, m.poster_url, m.backdrop_url, m.imdb_rating
			FROM movies m
			JOIN movie_category_mapping mcm ON m.id = mcm.movie_id
			WHERE mcm.category_id IN (
				SELECT mcm2.category_id 
				FROM user_watch_history uwh2 
				JOIN movie_category_mapping mcm2 ON uwh2.movie_id = mcm2.movie_id 
				WHERE uwh2.profile_id = ?
			)
			AND m.id NOT IN (SELECT movie_id FROM user_watch_history WHERE profile_id = ?)
			ORDER BY m.imdb_rating DESC
			LIMIT 10
		`
		rows, err = config.DB.Query(query, profileID, profileID)
	} else {
		query := `
			SELECT DISTINCT m.id, m.title, m.description, m.poster_url, m.backdrop_url, m.imdb_rating
			FROM movies m
			JOIN movie_category_mapping mcm ON m.id = mcm.movie_id
			WHERE mcm.category_id IN (
				SELECT mcm2.category_id 
				FROM user_watch_history uwh2 
				JOIN movie_category_mapping mcm2 ON uwh2.movie_id = mcm2.movie_id 
				WHERE uwh2.user_id = ?
			)
			AND m.id NOT IN (SELECT movie_id FROM user_watch_history WHERE user_id = ?)
			ORDER BY m.imdb_rating DESC
			LIMIT 10
		`
		rows, err = config.DB.Query(query, userID, userID)
	}

	if err != nil {
		http.Error(w, `{"error": "Failed to fetch recommendations"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var m models.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.PosterURL, &m.BackdropURL, &m.IMDBRating); err == nil {
			movies = append(movies, m)
		}
	}

	if len(movies) == 0 {
		// Fallback: Just return highest rated overall
		fallbackQuery := `SELECT id, title, description, poster_url, backdrop_url, imdb_rating FROM movies ORDER BY imdb_rating DESC LIMIT 10`
		fbRows, _ := config.DB.Query(fallbackQuery)
		if fbRows != nil {
			defer fbRows.Close()
			for fbRows.Next() {
				var m models.Movie
				if err := fbRows.Scan(&m.ID, &m.Title, &m.Description, &m.PosterURL, &m.BackdropURL, &m.IMDBRating); err == nil {
					movies = append(movies, m)
				}
			}
		}
	}

	if movies == nil {
		movies = []models.Movie{}
	}

	recommendationCache.Store(cacheKey, CachedRecommendation{
		Movies:    movies,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	})

	json.NewEncoder(w).Encode(movies)
}
