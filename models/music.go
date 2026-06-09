package models

import (
	"encoding/json"
	"time"
)

type MusicContent struct {
	ID           int             `json:"id"`
	Title        string          `json:"title"`
	Description  *string         `json:"description"`
	Artist       *string         `json:"artist"`
	VideoSources json.RawMessage `json:"video_sources"`
	AudioSources json.RawMessage `json:"audio_sources"`
	PosterURL    *string         `json:"poster_url"`
	BackdropURL  *string         `json:"backdrop_url"`
	ReleaseDate  *string         `json:"release_date"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
