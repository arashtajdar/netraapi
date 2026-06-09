package services

import (
	"context"
	"testing"

	"sheedbox-api/models"
)

type mockSeriesRepo struct {
	seriesList []models.Series
	seasons    map[int][]models.Season   // seriesID -> seasons
	episodes   map[int][]models.Episode  // seasonID -> episodes
}

func (m *mockSeriesRepo) List(ctx context.Context) ([]models.Series, error) {
	return m.seriesList, nil
}

func (m *mockSeriesRepo) GetByID(ctx context.Context, id int) (*models.Series, error) {
	for _, s := range m.seriesList {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, nil
}

func (m *mockSeriesRepo) GetSeasonsBySeriesID(ctx context.Context, seriesID int) ([]models.Season, error) {
	return m.seasons[seriesID], nil
}

func (m *mockSeriesRepo) GetEpisodesBySeasonID(ctx context.Context, seasonID int) ([]models.Episode, error) {
	return m.episodes[seasonID], nil
}

func (m *mockSeriesRepo) Create(ctx context.Context, s *models.Series) error {
	m.seriesList = append(m.seriesList, *s)
	return nil
}

func TestSeriesService_ListSeries(t *testing.T) {
	mock := &mockSeriesRepo{
		seriesList: []models.Series{
			{ID: 1, Title: "Test Series 1"},
		},
	}
	service := NewSeriesService(mock)
	res, err := service.ListSeries(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(res) != 1 || res[0].Title != "Test Series 1" {
		t.Errorf("expected Test Series 1, got %v", res)
	}
}

func TestSeriesService_GetSeriesDetail(t *testing.T) {
	mock := &mockSeriesRepo{
		seriesList: []models.Series{
			{ID: 1, Title: "Test Series 1"},
		},
		seasons: map[int][]models.Season{
			1: {
				{ID: 10, SeriesID: 1, SeasonNumber: 1},
			},
		},
		episodes: map[int][]models.Episode{
			10: {
				{ID: 100, SeasonID: 10, EpisodeNumber: 1, VideoSources: []byte(`[{"quality":"1080p","url":"http://test.com"}]`)},
			},
		},
	}

	service := NewSeriesService(mock)
	detail, err := service.GetSeriesDetail(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if detail == nil {
		t.Fatal("expected series detail to be found, got nil")
	}

	if detail.Title != "Test Series 1" {
		t.Errorf("expected Title 'Test Series 1', got '%s'", detail.Title)
	}

	if len(detail.Seasons) != 1 {
		t.Fatalf("expected 1 season, got %d", len(detail.Seasons))
	}

	season := detail.Seasons[0]
	if season.ID != 10 {
		t.Errorf("expected season ID 10, got %d", season.ID)
	}

	if len(season.Episodes) != 1 {
		t.Fatalf("expected 1 episode, got %d", len(season.Episodes))
	}

	ep := season.Episodes[0]
	if ep.ID != 100 {
		t.Errorf("expected episode ID 100, got %d", ep.ID)
	}
}

func TestSeriesService_CreateSeries(t *testing.T) {
	mock := &mockSeriesRepo{}
	service := NewSeriesService(mock)

	newSeries := &models.Series{Title: "New Show"}
	err := service.CreateSeries(context.Background(), newSeries)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(mock.seriesList) != 1 || mock.seriesList[0].Title != "New Show" {
		t.Errorf("expected series list to contain 'New Show', got %v", mock.seriesList)
	}
}
