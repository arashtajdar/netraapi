package repository

import (
	"context"
	"sheedbox-api/models"
)

// MovieRepository defines the data access methods for the Movie domain.
type MovieRepository interface {
	List(ctx context.Context) ([]models.Movie, error)
	GetByID(ctx context.Context, id int) (*models.Movie, error)
	Create(ctx context.Context, m *models.Movie) error
}

// SeriesRepository defines the data access methods for the Series domain.
type SeriesRepository interface {
	List(ctx context.Context) ([]models.Series, error)
	GetByID(ctx context.Context, id int) (*models.Series, error)
	GetSeasonsBySeriesID(ctx context.Context, seriesID int) ([]models.Season, error)
	GetEpisodesBySeasonID(ctx context.Context, seasonID int) ([]models.Episode, error)
	Create(ctx context.Context, s *models.Series) error
}

// LiveTVRepository defines the data access methods for the Live TV domain.
type LiveTVRepository interface {
	List(ctx context.Context) ([]models.LiveTVChannel, error)
	GetByID(ctx context.Context, id int) (*models.LiveTVChannel, error)
	GetEPGForChannel(ctx context.Context, channelID int) ([]models.EPG, error)
	Create(ctx context.Context, c *models.LiveTVChannel) (int64, error)
	Update(ctx context.Context, c *models.LiveTVChannel) error
	Delete(ctx context.Context, id int) error
	SaveEPG(ctx context.Context, channelID int64, epg []models.EPG) error
	UpdateYoutubeURL(ctx context.Context, id int, url string) error
}

// SportsRepository defines the data access methods for the Sports domain.
type SportsRepository interface {
	List(ctx context.Context) ([]models.SportsEvent, error)
	Create(ctx context.Context, e *models.SportsEvent) error
}

// MusicRepository defines the data access methods for the Music domain.
type MusicRepository interface {
	List(ctx context.Context) ([]models.MusicContent, error)
	GetByID(ctx context.Context, id int) (*models.MusicContent, error)
	Create(ctx context.Context, m *models.MusicContent) error
	Delete(ctx context.Context, id int) error
}

// FeaturedRepository defines data access methods for featured content banners.
type FeaturedRepository interface {
	List(ctx context.Context) ([]map[string]interface{}, error)
}

// UserProfileRepository defines data access methods for the User Profile domain.
type UserProfileRepository interface {
	GetByUserID(ctx context.Context, userID int) ([]models.UserProfile, error)
	Create(ctx context.Context, p *models.UserProfile) error
	Update(ctx context.Context, p *models.UserProfile) error
	Delete(ctx context.Context, id int, userID int) error
}

// WatchlistRepository defines data access methods for the Watchlist domain.
type WatchlistRepository interface {
	GetByUserID(ctx context.Context, userID int) ([]models.Watchlist, error)
}

// UserRepository defines data access methods for the User domain.
type UserRepository interface {
	GetByID(ctx context.Context, id int) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, string, error) // returns user and password hash
	Create(ctx context.Context, u *models.User) error
	UpdateCoins(ctx context.Context, userID int, coins int) error
}




