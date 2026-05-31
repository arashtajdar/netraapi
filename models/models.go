package models

import (
	"encoding/json"
	"time"
)

type User struct {
	ID           int             `json:"id"`
	Username     string          `json:"username"`
	Email        string          `json:"email"`
	PasswordHash string          `json:"-"`
	VirtualCoins int             `json:"virtual_coins"`
	ProfileData  json.RawMessage `json:"profile_data"`
	Settings     json.RawMessage `json:"settings"`
	CreatedAt    time.Time       `json:"created_at"`
}

type Watchlist struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

type WatchlistItem struct {
	WatchlistID int       `json:"watchlist_id"`
	ContentID   int       `json:"content_id"`
	ContentType string    `json:"content_type"`
	AddedAt     time.Time `json:"added_at"`
}

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

type LiveTVChannel struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	StreamURL  *string   `json:"stream_url"`
	LogoURL    *string   `json:"logo_url"`
	YoutubeURL *string   `json:"youtube_url"`
	CreatedAt  time.Time `json:"created_at"`
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

type SportsEvent struct {
	ID            int             `json:"id"`
	Title         string          `json:"title"`
	Description   *string         `json:"description"`
	IsLive        bool            `json:"is_live"`
	LiveStreamURL *string         `json:"live_stream_url"`
	VideoSources  json.RawMessage `json:"video_sources"`
	StartTime     *time.Time      `json:"start_time"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type UserWatchHistory struct {
	UserID                int       `json:"user_id"`
	MovieID               int       `json:"movie_id"`
	ResumePositionSeconds int       `json:"resume_position_seconds"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type WatchPartyRoom struct {
	ID                     int       `json:"id"`
	RoomCode               string    `json:"room_code"`
	HostID                 int       `json:"host_id"`
	CurrentMovieID         *int      `json:"current_movie_id,omitempty"`
	IsPlaying              bool      `json:"is_playing"`
	CurrentPositionSeconds int       `json:"current_position_seconds"`
	CreatedAt              time.Time `json:"created_at"`
}
