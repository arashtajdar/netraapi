package services

import (
	"context"
	"sheedbox-api/models"
	"sheedbox-api/repository"
)

type ProfileService struct {
	repo repository.UserProfileRepository
}

func NewProfileService(repo repository.UserProfileRepository) *ProfileService {
	return &ProfileService{repo: repo}
}

func (s *ProfileService) GetProfiles(ctx context.Context, userID int) ([]models.UserProfile, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *ProfileService) CreateProfile(ctx context.Context, p *models.UserProfile) error {
	return s.repo.Create(ctx, p)
}

func (s *ProfileService) UpdateProfile(ctx context.Context, p *models.UserProfile) error {
	return s.repo.Update(ctx, p)
}

func (s *ProfileService) DeleteProfile(ctx context.Context, id int, userID int) error {
	return s.repo.Delete(ctx, id, userID)
}
