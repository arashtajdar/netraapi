package services

import (
	"context"
	"sheedbox-api/models"
	"sheedbox-api/repository"
)

// SeriesService handles business logic for the Series domain.
type SeriesService struct {
	repo repository.SeriesRepository
}

// SeasonDetail embeds a season with its episodes.
type SeasonDetail struct {
	models.Season
	Episodes []models.Episode `json:"episodes"`
}

// SeriesDetail aggregates a series with all its seasons and episodes.
type SeriesDetail struct {
	models.Series
	Seasons []SeasonDetail `json:"seasons"`
}

// NewSeriesService creates a new SeriesService.
func NewSeriesService(repo repository.SeriesRepository) *SeriesService {
	return &SeriesService{repo: repo}
}

// ListSeries retrieves the list of series.
func (s *SeriesService) ListSeries(ctx context.Context) ([]models.Series, error) {
	return s.repo.List(ctx)
}

// GetSeriesDetail retrieves a series, its seasons, and episodes, applying logic like URL signing.
func (s *SeriesService) GetSeriesDetail(ctx context.Context, id int) (*SeriesDetail, error) {
	series, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if series == nil {
		return nil, nil // Not found
	}

	seasons, err := s.repo.GetSeasonsBySeriesID(ctx, series.ID)
	if err != nil {
		return nil, err
	}

	var seasonDetails []SeasonDetail
	for _, season := range seasons {
		episodes, err := s.repo.GetEpisodesBySeasonID(ctx, season.ID)
		if err != nil {
			// Alternatively return error, but usually skip or log
			continue 
		}

		// Apply business logic: sign video sources
		for i := range episodes {
			if episodes[i].VideoSources != nil {
				episodes[i].VideoSources = SignVideoSources(episodes[i].VideoSources)
			}
		}

		seasonDetails = append(seasonDetails, SeasonDetail{
			Season:   season,
			Episodes: episodes,
		})
	}

	return &SeriesDetail{
		Series:  *series,
		Seasons: seasonDetails,
	}, nil
}

func (s *SeriesService) CreateSeries(ctx context.Context, series *models.Series) error {
	return s.repo.Create(ctx, series)
}
