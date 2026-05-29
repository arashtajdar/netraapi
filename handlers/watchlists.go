package handlers

import (
	"encoding/json"
	"net/http"
	"netra-api/config"
	"netra-api/models"
)

func GetUserWatchlists(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := r.Context().Value("user_id").(int)

	query := `SELECT id, user_id, name, is_default, created_at FROM watchlists WHERE user_id = ?`
	rows, err := config.DB.Query(query, userID)
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var lists []models.Watchlist
	for rows.Next() {
		var wl models.Watchlist
		if err := rows.Scan(&wl.ID, &wl.UserID, &wl.Name, &wl.IsDefault, &wl.CreatedAt); err == nil {
			lists = append(lists, wl)
		}
	}
	if lists == nil {
		lists = []models.Watchlist{}
	}
	json.NewEncoder(w).Encode(lists)
}
