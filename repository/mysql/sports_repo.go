package mysql

import (
	"context"
	"database/sql"

	"sheedbox-api/contextkeys"
	"sheedbox-api/models"
	"sheedbox-api/repository"
)

type sportsRepo struct {
	db *sql.DB
}

// NewSportsRepository creates a new MySQL implementation of SportsRepository.
func NewSportsRepository(db *sql.DB) repository.SportsRepository {
	return &sportsRepo{db: db}
}

func (r *sportsRepo) List(ctx context.Context) ([]models.SportsEvent, error) {
	userLevel := contextkeys.UserLevelFromContext(ctx)
	query := `SELECT id, title, description, is_live, live_stream_url, video_sources, start_time, access_level, created_at, updated_at FROM sports_events WHERE access_level <= ?`
	rows, err := r.db.QueryContext(ctx, query, userLevel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.SportsEvent
	for rows.Next() {
		var e models.SportsEvent
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.IsLive, &e.LiveStreamURL, &e.VideoSources, &e.StartTime, &e.AccessLevel, &e.CreatedAt, &e.UpdatedAt); err == nil {
			events = append(events, e)
		}
	}
	if events == nil {
		events = []models.SportsEvent{}
	}
	return events, nil
}

func (r *sportsRepo) Create(ctx context.Context, e *models.SportsEvent) error {
	query := `INSERT INTO sports_events (title, description, is_live, live_stream_url, video_sources, start_time, access_level) 
		VALUES (?, NULLIF(?,''), ?, NULLIF(?,''), ?, NULLIF(?,''), ?)`
	res, err := r.db.ExecContext(ctx, query, e.Title, e.Description, e.IsLive, e.LiveStreamURL, e.VideoSources, e.StartTime, e.AccessLevel)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		e.ID = int(id)
	}
	return nil
}

