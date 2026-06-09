package services

import (
	"context"
	"sheedbox-api/models"
	"sheedbox-api/repository"
)

type MusicService struct {
	repo repository.MusicRepository
}

func NewMusicService(repo repository.MusicRepository) *MusicService {
	return &MusicService{repo: repo}
}

func (s *MusicService) ListMusic(ctx context.Context) ([]models.MusicContent, error) {
	return s.repo.List(ctx)
}

func (s *MusicService) GetMusicDetail(ctx context.Context, id int) (*models.MusicContent, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, nil
	}
	if m.VideoSources != nil {
		m.VideoSources = SignVideoSources(m.VideoSources)
	}
	if m.AudioSources != nil {
		m.AudioSources = SignVideoSources(m.AudioSources)
	}
	return m, nil
}

// CreateMusic handles creating a new music track.
func (s *MusicService) CreateMusic(ctx context.Context, m *models.MusicContent) error {
	return s.repo.Create(ctx, m)
}

// DeleteMusic handles deleting a music track.
func (s *MusicService) DeleteMusic(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

