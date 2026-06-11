package mysql

import (
	"context"
	"database/sql"
	"time"

	"sheedbox-api/contextkeys"
	"sheedbox-api/models"
	"sheedbox-api/repository"
)

type livetvRepo struct {
	db *sql.DB
}

// NewLiveTVRepository creates a new MySQL implementation of LiveTVRepository.
func NewLiveTVRepository(db *sql.DB) repository.LiveTVRepository {
	return &livetvRepo{db: db}
}

func (r *livetvRepo) List(ctx context.Context) ([]models.LiveTVChannel, error) {
	userLevel := contextkeys.UserLevelFromContext(ctx)
	query := `SELECT id, name, slug, stream_url, logo_url, youtube_url, youtube_channel_url, epg_fetch_url, last_epg_fetch, next_epg_fetch, access_level, created_at FROM live_tv_channels WHERE access_level <= ?`
	rows, err := r.db.QueryContext(ctx, query, userLevel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []models.LiveTVChannel
	for rows.Next() {
		var c models.LiveTVChannel
		var lastEPG, nextEPG []byte
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.StreamURL, &c.LogoURL, &c.YoutubeURL, &c.YoutubeChannelURL, &c.EPGFetchURL, &lastEPG, &nextEPG, &c.AccessLevel, &c.CreatedAt); err == nil {
			if len(lastEPG) > 0 {
				if t, err := time.Parse("2006-01-02 15:04:05", string(lastEPG)); err == nil {
					c.LastEPGFetch = &t
				} else if t, err := time.Parse(time.RFC3339, string(lastEPG)); err == nil {
					c.LastEPGFetch = &t
				}
			}
			if len(nextEPG) > 0 {
				if t, err := time.Parse("2006-01-02 15:04:05", string(nextEPG)); err == nil {
					c.NextEPGFetch = &t
				} else if t, err := time.Parse(time.RFC3339, string(nextEPG)); err == nil {
					c.NextEPGFetch = &t
				}
			}
			channels = append(channels, c)
		}
	}
	if channels == nil {
		channels = []models.LiveTVChannel{}
	}
	return channels, nil
}

func (r *livetvRepo) GetByID(ctx context.Context, id int) (*models.LiveTVChannel, error) {
	query := `SELECT id, name, slug, stream_url, logo_url, youtube_url, youtube_channel_url, epg_fetch_url, last_epg_fetch, next_epg_fetch, access_level, created_at FROM live_tv_channels WHERE id = ?`
	var c models.LiveTVChannel
	var lastEPG, nextEPG []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(&c.ID, &c.Name, &c.Slug, &c.StreamURL, &c.LogoURL, &c.YoutubeURL, &c.YoutubeChannelURL, &c.EPGFetchURL, &lastEPG, &nextEPG, &c.AccessLevel, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(lastEPG) > 0 {
		if t, err := time.Parse("2006-01-02 15:04:05", string(lastEPG)); err == nil {
			c.LastEPGFetch = &t
		} else if t, err := time.Parse(time.RFC3339, string(lastEPG)); err == nil {
			c.LastEPGFetch = &t
		}
	}
	if len(nextEPG) > 0 {
		if t, err := time.Parse("2006-01-02 15:04:05", string(nextEPG)); err == nil {
			c.NextEPGFetch = &t
		} else if t, err := time.Parse(time.RFC3339, string(nextEPG)); err == nil {
			c.NextEPGFetch = &t
		}
	}
	return &c, nil
}

func (r *livetvRepo) GetEPGForChannel(ctx context.Context, channelID int) ([]models.EPG, error) {
	query := `SELECT id, channel_id, program_title, description, start_time, end_time, created_at FROM epg WHERE channel_id = ? ORDER BY start_time ASC`
	rows, err := r.db.QueryContext(ctx, query, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var epgs []models.EPG
	for rows.Next() {
		var e models.EPG
		var start, end []byte
		var created []byte
		if err := rows.Scan(&e.ID, &e.ChannelID, &e.ProgramTitle, &e.Description, &start, &end, &created); err == nil {
			sStart := string(start)
			sEnd := string(end)
			sCreated := string(created)

			if t, err := time.Parse("2006-01-02 15:04:05", sStart); err == nil {
				e.StartTime = t
			} else if t, err := time.Parse(time.RFC3339, sStart); err == nil {
				e.StartTime = t
			}

			if t, err := time.Parse("2006-01-02 15:04:05", sEnd); err == nil {
				e.EndTime = t
			} else if t, err := time.Parse(time.RFC3339, sEnd); err == nil {
				e.EndTime = t
			}

			if t, err := time.Parse("2006-01-02 15:04:05", sCreated); err == nil {
				e.CreatedAt = t
			} else if t, err := time.Parse(time.RFC3339, sCreated); err == nil {
				e.CreatedAt = t
			}
			epgs = append(epgs, e)
		}
	}
	if epgs == nil {
		epgs = []models.EPG{}
	}
	return epgs, nil
}

func (r *livetvRepo) Create(ctx context.Context, c *models.LiveTVChannel) (int64, error) {
	query := `INSERT INTO live_tv_channels (name, slug, stream_url, logo_url, youtube_url, youtube_channel_url, epg_fetch_url, last_epg_fetch, next_epg_fetch, access_level) VALUES (?, ?, NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, query, c.Name, c.Slug, c.StreamURL, c.LogoURL, c.YoutubeURL, c.YoutubeChannelURL, c.EPGFetchURL, c.LastEPGFetch, c.NextEPGFetch, c.AccessLevel)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err == nil {
		c.ID = int(id)
	}
	return id, nil
}

func (r *livetvRepo) Update(ctx context.Context, c *models.LiveTVChannel) error {
	query := `UPDATE live_tv_channels SET name=?, slug=?, stream_url=NULLIF(?,''), logo_url=NULLIF(?,''), youtube_url=NULLIF(?,''), youtube_channel_url=NULLIF(?,''), epg_fetch_url=NULLIF(?,''), last_epg_fetch=?, next_epg_fetch=?, access_level=? WHERE id=?`
	_, err := r.db.ExecContext(ctx, query, c.Name, c.Slug, c.StreamURL, c.LogoURL, c.YoutubeURL, c.YoutubeChannelURL, c.EPGFetchURL, c.LastEPGFetch, c.NextEPGFetch, c.AccessLevel, c.ID)
	return err
}

func (r *livetvRepo) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM live_tv_channels WHERE id = ?", id)
	return err
}

func (r *livetvRepo) SaveEPG(ctx context.Context, channelID int64, epg []models.EPG) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM epg WHERE channel_id = ?", channelID)
	if err != nil {
		return err
	}
	for _, item := range epg {
		_, err = r.db.ExecContext(ctx, "INSERT INTO epg (channel_id, program_title, description, start_time, end_time) VALUES (?, ?, ?, ?, ?)",
			channelID, item.ProgramTitle, item.Description, item.StartTime, item.EndTime)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *livetvRepo) UpdateYoutubeURL(ctx context.Context, id int, url string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE live_tv_channels SET youtube_url=? WHERE id=?", url, id)
	return err
}

