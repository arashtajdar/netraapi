package models

import "time"

// User represents the users table
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Hidden from JSON responses
	VirtualCoins int       `json:"virtual_coins"`
	CreatedAt    time.Time `json:"created_at"`
}

// Movie represents the movies table
type Movie struct {
	ID              int    `json:"id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	VideoURL        string `json:"video_url"`
	ThumbnailURL    string `json:"thumbnail_url"`
	Category        string `json:"category"`
	DurationSeconds int    `json:"duration_seconds"`
}

// UserWatchHistory represents the user_watch_history table
type UserWatchHistory struct {
	UserID                int       `json:"user_id"`
	MovieID               int       `json:"movie_id"`
	ResumePositionSeconds int       `json:"resume_position_seconds"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// WatchPartyRoom represents the watch_party_rooms table
type WatchPartyRoom struct {
	ID                     int       `json:"id"`
	RoomCode               string    `json:"room_code"`
	HostID                 int       `json:"host_id"`
	CurrentMovieID         *int      `json:"current_movie_id,omitempty"`
	IsPlaying              bool      `json:"is_playing"`
	CurrentPositionSeconds int       `json:"current_position_seconds"`
	CreatedAt              time.Time `json:"created_at"`
}
