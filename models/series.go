package models

import (
	"encoding/json"
	"time"
)

type Series struct {
	ID          int             `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Director    *string         `json:"director"`
	CastMembers json.RawMessage `json:"cast_members"`
	Rating      *float64        `json:"rating"`
	PosterURL   *string         `json:"poster_url"`
	BackdropURL *string         `json:"backdrop_url"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type Season struct {
	ID           int       `json:"id"`
	SeriesID     int       `json:"series_id"`
	SeasonNumber int       `json:"season_number"`
	Title        *string   `json:"title"`
	Description  *string   `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
}

type Episode struct {
	ID            int             `json:"id"`
	SeasonID      int             `json:"season_id"`
	EpisodeNumber int             `json:"episode_number"`
	Title         *string         `json:"title"`
	Description   *string         `json:"description"`
	VideoSources  json.RawMessage `json:"video_sources"`
	Subtitles     json.RawMessage `json:"subtitles"`
	IntroStart    *int            `json:"intro_start"`
	IntroEnd      *int            `json:"intro_end"`
	CreatedAt     time.Time       `json:"created_at"`
}
