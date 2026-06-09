package mysql

import (
	"context"
	"database/sql"

	"sheedbox-api/models"
	"sheedbox-api/repository"
)

type watchlistRepo struct {
	db *sql.DB
}

// NewWatchlistRepository creates a new MySQL implementation of WatchlistRepository.
func NewWatchlistRepository(db *sql.DB) repository.WatchlistRepository {
	return &watchlistRepo{db: db}
}

func (r *watchlistRepo) GetByUserID(ctx context.Context, userID int) ([]models.Watchlist, error) {
	query := `SELECT id, user_id, name, is_default, created_at FROM watchlists WHERE user_id = ?`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []models.Watchlist
	for rows.Next() {
		var wl models.Watchlist
		if err := rows.Scan(&wl.ID, &wl.UserID, &wl.Name, &wl.IsDefault, &wl.CreatedAt); err == nil {
			lists = append(lists, wl)
		}
	}
	if lists == nil {
		lists = []models.Watchlist{}
	}
	return lists, nil
}
