package models

import (
	"encoding/json"
	"time"
)

type SportsEvent struct {
	ID            int             `json:"id"`
	Title         string          `json:"title"`
	Description   *string         `json:"description"`
	IsLive        bool            `json:"is_live"`
	LiveStreamURL *string         `json:"live_stream_url"`
	VideoSources  json.RawMessage `json:"video_sources"`
	StartTime     *time.Time      `json:"start_time"`
	AccessLevel   int             `json:"access_level"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}
