package mysql

import (
	"context"
	"database/sql"

	"sheedbox-api/contextkeys"
	"sheedbox-api/models"
	"sheedbox-api/repository"
)

type movieRepo struct {
	db *sql.DB
}

// NewMovieRepository creates a new MySQL implementation of MovieRepository.
func NewMovieRepository(db *sql.DB) repository.MovieRepository {
	return &movieRepo{db: db}
}

func (r *movieRepo) List(ctx context.Context) ([]models.Movie, error) {
	userLevel := contextkeys.UserLevelFromContext(ctx)
	query := `SELECT id, title, description, release_date, director, cast_members, imdb_rating, local_rating, poster_url, backdrop_url, access_level, created_at, updated_at FROM movies WHERE access_level <= ? ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, userLevel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var m models.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.ReleaseDate, &m.Director, &m.CastMembers, &m.IMDBRating, &m.LocalRating, &m.PosterURL, &m.BackdropURL, &m.AccessLevel, &m.CreatedAt, &m.UpdatedAt); err == nil {
			movies = append(movies, m)
		}
	}
	if movies == nil {
		movies = []models.Movie{}
	}
	return movies, nil
}

func (r *movieRepo) GetByID(ctx context.Context, id int) (*models.Movie, error) {
	query := `SELECT id, title, description, release_date, director, cast_members, imdb_rating, local_rating, poster_url, backdrop_url, video_sources, subtitles, intro_start, intro_end, access_level, created_at, updated_at FROM movies WHERE id = ?`
	
	var m models.Movie
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.Title, &m.Description, &m.ReleaseDate, &m.Director, &m.CastMembers, 
		&m.IMDBRating, &m.LocalRating, &m.PosterURL, &m.BackdropURL, 
		&m.VideoSources, &m.Subtitles, &m.IntroStart, &m.IntroEnd, &m.AccessLevel,
		&m.CreatedAt, &m.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil // Return nil, nil to indicate not found clearly
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *movieRepo) Create(ctx context.Context, m *models.Movie) error {
	query := `INSERT INTO movies (title, description, release_date, director, imdb_rating, local_rating, poster_url, backdrop_url, intro_start, intro_end, access_level, video_sources, subtitles) 
			  VALUES (?, ?, NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), ?, ?, ?)`

	res, err := r.db.ExecContext(ctx, query, 
		m.Title, m.Description, m.ReleaseDate, m.Director, m.IMDBRating, m.LocalRating, 
		m.PosterURL, m.BackdropURL, m.IntroStart, m.IntroEnd, m.AccessLevel, m.VideoSources, m.Subtitles)
	
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err == nil {
		m.ID = int(id)
	}

	return nil
}
