package services

import (
	"context"

	"sheedbox-api/models"
	"sheedbox-api/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, string, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *UserService) CreateUser(ctx context.Context, u *models.User) error {
	return s.repo.Create(ctx, u)
}

func (s *UserService) AwardCoins(ctx context.Context, userID int, coins int) error {
	return s.repo.UpdateCoins(ctx, userID, coins)
}
