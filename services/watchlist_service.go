package services

import (
	"context"
	"sheedbox-api/models"
	"sheedbox-api/repository"
)

type WatchlistService struct {
	repo repository.WatchlistRepository
}

func NewWatchlistService(repo repository.WatchlistRepository) *WatchlistService {
	return &WatchlistService{repo: repo}
}

func (s *WatchlistService) GetWatchlists(ctx context.Context, userID int) ([]models.Watchlist, error) {
	return s.repo.GetByUserID(ctx, userID)
}
