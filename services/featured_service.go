package services

import (
	"context"
	"sheedbox-api/repository"
)

type FeaturedService struct {
	repo repository.FeaturedRepository
}

func NewFeaturedService(repo repository.FeaturedRepository) *FeaturedService {
	return &FeaturedService{repo: repo}
}

func (s *FeaturedService) ListFeatured(ctx context.Context) ([]map[string]interface{}, error) {
	return s.repo.List(ctx)
}
