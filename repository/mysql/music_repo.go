package mysql

import (
	"context"
	"database/sql"

	"sheedbox-api/models"
	"sheedbox-api/repository"
)

type musicRepo struct {
	db *sql.DB
}

// NewMusicRepository creates a new MySQL implementation.
func NewMusicRepository(db *sql.DB) repository.MusicRepository {
	return &musicRepo{db: db}
}

func (r *musicRepo) List(ctx context.Context) ([]models.MusicContent, error) {
	query := `SELECT id, title, description, artist, poster_url, backdrop_url, created_at, updated_at FROM music_content`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var music []models.MusicContent
	for rows.Next() {
		var m models.MusicContent
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.Artist, &m.PosterURL, &m.BackdropURL, &m.CreatedAt, &m.UpdatedAt); err == nil {
			music = append(music, m)
		}
	}
	if music == nil {
		music = []models.MusicContent{}
	}
	return music, nil
}

func (r *musicRepo) GetByID(ctx context.Context, id int) (*models.MusicContent, error) {
	query := `SELECT id, title, description, artist, video_sources, audio_sources, poster_url, backdrop_url, release_date, created_at, updated_at FROM music_content WHERE id = ?`
	var m models.MusicContent
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.Title, &m.Description, &m.Artist, &m.VideoSources, &m.AudioSources, 
		&m.PosterURL, &m.BackdropURL, &m.ReleaseDate, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *musicRepo) Create(ctx context.Context, m *models.MusicContent) error {
	query := `INSERT INTO music_content (title, description, artist, release_date, poster_url, backdrop_url, video_sources) 
		VALUES (?, NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), ?)`
	res, err := r.db.ExecContext(ctx, query, m.Title, m.Description, m.Artist, m.ReleaseDate, m.PosterURL, m.BackdropURL, m.VideoSources)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		m.ID = int(id)
	}
	return nil
}

func (r *musicRepo) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM music_content WHERE id = ?", id)
	return err
}

