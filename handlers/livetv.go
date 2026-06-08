package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"sheedbox-api/config"
	"sheedbox-api/models"
)

func GetLiveChannels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := `SELECT id, name, stream_url, logo_url, youtube_url, created_at FROM live_tv_channels`
	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var channels []models.LiveTVChannel
	for rows.Next() {
		var c models.LiveTVChannel
		if err := rows.Scan(&c.ID, &c.Name, &c.StreamURL, &c.LogoURL, &c.YoutubeURL, &c.CreatedAt); err == nil {
			c.EPG = fetchEPGForChannel(c.ID)
			channels = append(channels, c)
		}
	}
	if channels == nil {
		channels = []models.LiveTVChannel{}
	}
	json.NewEncoder(w).Encode(channels)
}

func fetchEPGForChannel(channelID int) []models.EPG {
	rows, err := config.DB.Query("SELECT id, channel_id, program_title, description, start_time, end_time, created_at FROM epg WHERE channel_id = ? ORDER BY start_time ASC", channelID)
	if err != nil {
		return []models.EPG{}
	}
	defer rows.Close()

	var epgs []models.EPG
	for rows.Next() {
		var e models.EPG
		var start, end []byte
		var created []byte
		if err := rows.Scan(&e.ID, &e.ChannelID, &e.ProgramTitle, &e.Description, &start, &end, &created); err == nil {
			sStart := string(start)
			sEnd := string(end)
			sCreated := string(created)

			if t, err := time.Parse("2006-01-02 15:04:05", sStart); err == nil {
				e.StartTime = t
			} else if t, err := time.Parse(time.RFC3339, sStart); err == nil {
				e.StartTime = t
			}

			if t, err := time.Parse("2006-01-02 15:04:05", sEnd); err == nil {
				e.EndTime = t
			} else if t, err := time.Parse(time.RFC3339, sEnd); err == nil {
				e.EndTime = t
			}

			if t, err := time.Parse("2006-01-02 15:04:05", sCreated); err == nil {
				e.CreatedAt = t
			} else if t, err := time.Parse(time.RFC3339, sCreated); err == nil {
				e.CreatedAt = t
			}
			epgs = append(epgs, e)
		} else {
			// handle or ignore error
		}
	}
	if epgs == nil {
		return []models.EPG{}
	}
	return epgs
}
