package models

import (
	"encoding/json"
	"time"
)

type WatchPartyRoom struct {
	ID                     int       `json:"id"`
	RoomCode               string    `json:"room_code"`
	HostID                 int       `json:"host_id"`
	CurrentMovieID         *int      `json:"current_movie_id,omitempty"`
	IsPlaying              bool      `json:"is_playing"`
	CurrentPositionSeconds int       `json:"current_position_seconds"`
	CreatedAt              time.Time `json:"created_at"`
}

type AppSetting struct {
	SettingKey   string          `json:"setting_key"`
	SettingValue json.RawMessage `json:"setting_value"`
}
