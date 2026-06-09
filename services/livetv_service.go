package services

import (
	"context"
	"sheedbox-api/models"
	"sheedbox-api/repository"
)

// LiveTVService handles business logic for the Live TV domain.
type LiveTVService struct {
	repo repository.LiveTVRepository
}

// NewLiveTVService creates a new LiveTVService.
func NewLiveTVService(repo repository.LiveTVRepository) *LiveTVService {
	return &LiveTVService{repo: repo}
}

// ListChannels retrieves all live channels along with their EPG data.
func (s *LiveTVService) ListChannels(ctx context.Context) ([]models.LiveTVChannel, error) {
	channels, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	for i := range channels {
		// Sign stream URL
		if channels[i].StreamURL != nil {
			signed := SignURL(*channels[i].StreamURL)
			channels[i].StreamURL = &signed
		}

		// Fetch EPG data
		epg, err := s.repo.GetEPGForChannel(ctx, channels[i].ID)
		if err == nil {
			channels[i].EPG = epg
		} else {
			channels[i].EPG = []models.EPG{}
		}
	}

	return channels, nil
}

// GetChannel retrieves a single live channel.
func (s *LiveTVService) GetChannel(ctx context.Context, id int) (*models.LiveTVChannel, error) {
	return s.repo.GetByID(ctx, id)
}

// GetEPG retrieves EPG data for a channel.
func (s *LiveTVService) GetEPG(ctx context.Context, channelID int) ([]models.EPG, error) {
	return s.repo.GetEPGForChannel(ctx, channelID)
}

// CreateChannel creates a channel and returns the inserted ID.
func (s *LiveTVService) CreateChannel(ctx context.Context, c *models.LiveTVChannel) (int64, error) {
	return s.repo.Create(ctx, c)
}

// UpdateChannel updates the channel information.
func (s *LiveTVService) UpdateChannel(ctx context.Context, c *models.LiveTVChannel) error {
	return s.repo.Update(ctx, c)
}

// DeleteChannel deletes a live channel.
func (s *LiveTVService) DeleteChannel(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

// SaveEPG saves EPG data for a channel.
func (s *LiveTVService) SaveEPG(ctx context.Context, channelID int64, epg []models.EPG) error {
	return s.repo.SaveEPG(ctx, channelID, epg)
}

// UpdateYoutubeURL updates the live channel's youtube URL.
func (s *LiveTVService) UpdateYoutubeURL(ctx context.Context, id int, url string) error {
	return s.repo.UpdateYoutubeURL(ctx, id, url)
}

