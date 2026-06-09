package models

import (
	"encoding/json"
	"time"
)

type Movie struct {
	ID           int             `json:"id"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	ReleaseDate  *string         `json:"release_date"`
	Director     *string         `json:"director"`
	CastMembers  json.RawMessage `json:"cast_members"`
	IMDBRating   *float64        `json:"imdb_rating"`
	LocalRating  *float64        `json:"local_rating"`
	PosterURL    *string         `json:"poster_url"`
	BackdropURL  *string         `json:"backdrop_url"`
	VideoSources json.RawMessage `json:"video_sources"`
	Subtitles    json.RawMessage `json:"subtitles"`
	IntroStart   *int            `json:"intro_start"`
	IntroEnd     *int            `json:"intro_end"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
