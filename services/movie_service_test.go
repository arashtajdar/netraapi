package services

import (
	"context"
	"testing"

	"sheedbox-api/models"
)

// mockMovieRepo is a mock implementation of repository.MovieRepository
type mockMovieRepo struct {
	movies    []models.Movie
	getByIDFn func(id int) (*models.Movie, error)
	createFn  func(m *models.Movie) error
}

func (m *mockMovieRepo) List(ctx context.Context) ([]models.Movie, error) {
	return m.movies, nil
}

func (m *mockMovieRepo) GetByID(ctx context.Context, id int) (*models.Movie, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	for _, movie := range m.movies {
		if movie.ID == id {
			return &movie, nil
		}
	}
	return nil, nil
}

func (m *mockMovieRepo) Create(ctx context.Context, movie *models.Movie) error {
	if m.createFn != nil {
		return m.createFn(movie)
	}
	m.movies = append(m.movies, *movie)
	return nil
}

func TestMovieService_ListMovies(t *testing.T) {
	mockRepo := &mockMovieRepo{
		movies: []models.Movie{
			{ID: 1, Title: "Test Movie 1"},
			{ID: 2, Title: "Test Movie 2"},
		},
	}

	service := NewMovieService(mockRepo)
	res, err := service.ListMovies(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(res) != 2 {
		t.Errorf("expected 2 movies, got %d", len(res))
	}
}

func TestMovieService_GetMovieDetail(t *testing.T) {
	mockRepo := &mockMovieRepo{
		movies: []models.Movie{
			{ID: 1, Title: "Test Movie 1"},
		},
	}

	service := NewMovieService(mockRepo)

	// Test Found
	res, err := service.GetMovieDetail(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res == nil || res.Title != "Test Movie 1" {
		t.Errorf("expected Test Movie 1, got %v", res)
	}

	// Test Not Found
	res, err = service.GetMovieDetail(context.Background(), 999)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res != nil {
		t.Errorf("expected nil for non-existent movie, got %v", res)
	}
}
