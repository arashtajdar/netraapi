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
	UserLevel    int             `json:"user_level"`
	FirstName    *string         `json:"first_name"`
	LastName     *string         `json:"last_name"`
	AvatarURL    *string         `json:"avatar_url"`
	Settings     json.RawMessage `json:"settings"`
	CreatedAt    time.Time       `json:"created_at"`
}

type UserProfile struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	Name       string    `json:"name"`
	AvatarURL  *string   `json:"avatar_url"`
	IsKidsMode bool      `json:"is_kids_mode"`
	CreatedAt  time.Time `json:"created_at"`
}

type Watchlist struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ProfileID *int      `json:"profile_id"`
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

type UserWatchHistory struct {
	UserID                int       `json:"user_id"`
	ProfileID             *int      `json:"profile_id"`
	MovieID               int       `json:"movie_id"`
	ResumePositionSeconds int       `json:"resume_position_seconds"`
	UpdatedAt             time.Time `json:"updated_at"`
}
