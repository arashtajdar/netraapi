package services

import (
	"context"
	"sheedbox-api/models"
	"sheedbox-api/repository"
)

// MovieService handles business logic for the Movie domain.
type MovieService struct {
	repo repository.MovieRepository
}

// NewMovieService creates a new MovieService.
func NewMovieService(repo repository.MovieRepository) *MovieService {
	return &MovieService{repo: repo}
}

// ListMovies retrieves the public list of movies.
func (s *MovieService) ListMovies(ctx context.Context) ([]models.Movie, error) {
	return s.repo.List(ctx)
}

// GetMovieDetail retrieves a single movie and applies business logic (like URL signing).
func (s *MovieService) GetMovieDetail(ctx context.Context, id int) (*models.Movie, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, nil // Not found
	}

	// Apply business logic: Sign video sources if they exist
	if m.VideoSources != nil {
		m.VideoSources = SignVideoSources(m.VideoSources)
	}

	return m, nil
}

// CreateMovie handles creating a new movie (typically from the admin panel).
func (s *MovieService) CreateMovie(ctx context.Context, m *models.Movie) error {
	// Add any business validation here in the future
	return s.repo.Create(ctx, m)
}
