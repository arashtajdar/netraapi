package mysql

import (
	"context"
	"database/sql"

	"sheedbox-api/models"
	"sheedbox-api/repository"
)

type userRepo struct {
	db *sql.DB
}

// NewUserRepository creates a new MySQL implementation of UserRepository.
func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) GetByID(ctx context.Context, id int) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, virtual_coins, user_level, first_name, last_name, avatar_url, settings, created_at FROM users WHERE id = ?`
	var u models.User
	var passwordHash string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Username, &u.Email, &passwordHash, &u.VirtualCoins, &u.UserLevel,
		&u.FirstName, &u.LastName, &u.AvatarURL, &u.Settings, &u.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.PasswordHash = passwordHash
	return &u, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*models.User, string, error) {
	query := `SELECT id, username, email, password_hash, virtual_coins, user_level, first_name, last_name, avatar_url, settings, created_at FROM users WHERE email = ?`
	var u models.User
	var passwordHash string
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID, &u.Username, &u.Email, &passwordHash, &u.VirtualCoins, &u.UserLevel,
		&u.FirstName, &u.LastName, &u.AvatarURL, &u.Settings, &u.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	return &u, passwordHash, nil
}

func (r *userRepo) Create(ctx context.Context, u *models.User) error {
	query := `INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`
	res, err := r.db.ExecContext(ctx, query, u.Username, u.Email, u.PasswordHash)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		u.ID = int(id)
	}
	return nil
}

func (r *userRepo) UpdateCoins(ctx context.Context, userID int, coins int) error {
	query := `UPDATE users SET virtual_coins = virtual_coins + ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, coins, userID)
	return err
}
