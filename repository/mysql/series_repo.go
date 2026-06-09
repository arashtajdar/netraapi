package mysql

import (
	"context"
	"database/sql"

	"sheedbox-api/models"
	"sheedbox-api/repository"
)

type seriesRepo struct {
	db *sql.DB
}

// NewSeriesRepository creates a new MySQL implementation of SeriesRepository.
func NewSeriesRepository(db *sql.DB) repository.SeriesRepository {
	return &seriesRepo{db: db}
}

func (r *seriesRepo) List(ctx context.Context) ([]models.Series, error) {
	query := `SELECT id, title, description, director, cast_members, rating, poster_url, backdrop_url, created_at, updated_at FROM series`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seriesList []models.Series
	for rows.Next() {
		var s models.Series
		if err := rows.Scan(&s.ID, &s.Title, &s.Description, &s.Director, &s.CastMembers, &s.Rating, &s.PosterURL, &s.BackdropURL, &s.CreatedAt, &s.UpdatedAt); err == nil {
			seriesList = append(seriesList, s)
		}
	}
	if seriesList == nil {
		seriesList = []models.Series{}
	}
	return seriesList, nil
}

func (r *seriesRepo) GetByID(ctx context.Context, id int) (*models.Series, error) {
	query := `SELECT id, title, description, director, cast_members, rating, poster_url, backdrop_url, created_at, updated_at FROM series WHERE id = ?`
	var s models.Series
	err := r.db.QueryRowContext(ctx, query, id).Scan(&s.ID, &s.Title, &s.Description, &s.Director, &s.CastMembers, &s.Rating, &s.PosterURL, &s.BackdropURL, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *seriesRepo) GetSeasonsBySeriesID(ctx context.Context, seriesID int) ([]models.Season, error) {
	query := `SELECT id, series_id, season_number, title, description, created_at FROM seasons WHERE series_id = ? ORDER BY season_number ASC`
	rows, err := r.db.QueryContext(ctx, query, seriesID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seasonsList []models.Season
	for rows.Next() {
		var seas models.Season
		if err := rows.Scan(&seas.ID, &seas.SeriesID, &seas.SeasonNumber, &seas.Title, &seas.Description, &seas.CreatedAt); err == nil {
			seasonsList = append(seasonsList, seas)
		}
	}
	if seasonsList == nil {
		seasonsList = []models.Season{}
	}
	return seasonsList, nil
}

func (r *seriesRepo) GetEpisodesBySeasonID(ctx context.Context, seasonID int) ([]models.Episode, error) {
	query := `SELECT id, season_id, episode_number, title, description, video_sources, subtitles, intro_start, intro_end, created_at FROM episodes WHERE season_id = ? ORDER BY episode_number ASC`
	rows, err := r.db.QueryContext(ctx, query, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var episodes []models.Episode
	for rows.Next() {
		var ep models.Episode
		if err := rows.Scan(&ep.ID, &ep.SeasonID, &ep.EpisodeNumber, &ep.Title, &ep.Description, &ep.VideoSources, &ep.Subtitles, &ep.IntroStart, &ep.IntroEnd, &ep.CreatedAt); err == nil {
			episodes = append(episodes, ep)
		}
	}
	if episodes == nil {
		episodes = []models.Episode{}
	}
	return episodes, nil
}

func (r *seriesRepo) Create(ctx context.Context, s *models.Series) error {
	query := `INSERT INTO series (title, description, director, rating, poster_url, backdrop_url, cast_members) 
		VALUES (?, NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), ?)`
	res, err := r.db.ExecContext(ctx, query, s.Title, s.Description, s.Director, s.Rating, s.PosterURL, s.BackdropURL, s.CastMembers)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		s.ID = int(id)
	}
	return nil
}
