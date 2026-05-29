package handlers

import (
	"encoding/json"
	"net/http"
	"netra-api/config"
	"netra-api/models"
)

func GetSportsEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := `SELECT id, title, description, is_live, live_stream_url, video_sources, start_time, created_at, updated_at FROM sports_events`
	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []models.SportsEvent
	for rows.Next() {
		var e models.SportsEvent
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.IsLive, &e.LiveStreamURL, &e.VideoSources, &e.StartTime, &e.CreatedAt, &e.UpdatedAt); err == nil {
			events = append(events, e)
		}
	}
	if events == nil {
		events = []models.SportsEvent{}
	}
	json.NewEncoder(w).Encode(events)
}
