package handlers

import (
	"encoding/json"
	"net/http"
	"netra-api/config"
	"netra-api/models"
)

func GetLiveChannels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := `SELECT id, name, stream_url, logo_url, created_at FROM live_tv_channels`
	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var channels []models.LiveTVChannel
	for rows.Next() {
		var c models.LiveTVChannel
		if err := rows.Scan(&c.ID, &c.Name, &c.StreamURL, &c.LogoURL, &c.CreatedAt); err == nil {
			channels = append(channels, c)
		}
	}
	if channels == nil {
		channels = []models.LiveTVChannel{}
	}
	json.NewEncoder(w).Encode(channels)
}
