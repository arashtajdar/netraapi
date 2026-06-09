package services

import (
	"context"
	"sheedbox-api/models"
	"sheedbox-api/repository"
)

// SportsService handles business logic for the Sports domain.
type SportsService struct {
	repo repository.SportsRepository
}

// NewSportsService creates a new SportsService.
func NewSportsService(repo repository.SportsRepository) *SportsService {
	return &SportsService{repo: repo}
}

// ListEvents retrieves all sports events and signs their URLs.
func (s *SportsService) ListEvents(ctx context.Context) ([]models.SportsEvent, error) {
	events, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	for i := range events {
		if events[i].LiveStreamURL != nil {
			signed := SignURL(*events[i].LiveStreamURL)
			events[i].LiveStreamURL = &signed
		}
		if events[i].VideoSources != nil {
			events[i].VideoSources = SignVideoSources(events[i].VideoSources)
		}
	}

	return events, nil
}

// CreateEvent handles creating a new sports event.
func (s *SportsService) CreateEvent(ctx context.Context, e *models.SportsEvent) error {
	return s.repo.Create(ctx, e)
}

