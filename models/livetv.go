package models

import (
	"time"
)

type LiveTVChannel struct {
	ID                int       `json:"id"`
	Slug              string    `json:"slug"`
	Name              string    `json:"name"`
	StreamURL         *string   `json:"stream_url"`
	LogoURL           *string   `json:"logo_url"`
	YoutubeURL        *string   `json:"youtube_url"`
	YoutubeChannelURL *string   `json:"youtube_channel_url"`
	EPG               []EPG     `json:"epg"`
	CreatedAt         time.Time `json:"created_at"`
}

type EPG struct {
	ID           int       `json:"id"`
	ChannelID    int       `json:"channel_id"`
	ProgramTitle string    `json:"program_title"`
	Description  *string   `json:"description"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	CreatedAt    time.Time `json:"created_at"`
}
