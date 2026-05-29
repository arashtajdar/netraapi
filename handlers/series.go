package handlers

import (
	"encoding/json"
	"net/http"
	"netra-api/config"
	"netra-api/models"
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
