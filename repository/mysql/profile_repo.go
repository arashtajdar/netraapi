package mysql

import (
	"context"
	"database/sql"

	"sheedbox-api/models"
	"sheedbox-api/repository"
)

type userProfileRepo struct {
	db *sql.DB
}

// NewUserProfileRepository creates a new MySQL implementation of UserProfileRepository.
func NewUserProfileRepository(db *sql.DB) repository.UserProfileRepository {
	return &userProfileRepo{db: db}
}

func (r *userProfileRepo) GetByUserID(ctx context.Context, userID int) ([]models.UserProfile, error) {
	query := `SELECT id, user_id, name, avatar_url, is_kids_mode, created_at FROM user_profiles WHERE user_id = ?`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []models.UserProfile
	for rows.Next() {
		var p models.UserProfile
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.AvatarURL, &p.IsKidsMode, &p.CreatedAt); err == nil {
			profiles = append(profiles, p)
		}
	}
	if profiles == nil {
		profiles = []models.UserProfile{}
	}
	return profiles, nil
}

func (r *userProfileRepo) Create(ctx context.Context, p *models.UserProfile) error {
	query := `INSERT INTO user_profiles (user_id, name, avatar_url, is_kids_mode) VALUES (?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, query, p.UserID, p.Name, p.AvatarURL, p.IsKidsMode)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		p.ID = int(id)
	}
	return nil
}

func (r *userProfileRepo) Update(ctx context.Context, p *models.UserProfile) error {
	query := `UPDATE user_profiles SET name = ?, avatar_url = ?, is_kids_mode = ? WHERE id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, p.Name, p.AvatarURL, p.IsKidsMode, p.ID, p.UserID)
	return err
}

func (r *userProfileRepo) Delete(ctx context.Context, id int, userID int) error {
	query := `DELETE FROM user_profiles WHERE id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, id, userID)
	return err
}
