package repositories

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/mbnaysinger/fechou-brick-api/internal/core/domain"
)

type PostgresAdminRepository struct {
	db *sqlx.DB
}

func NewPostgresAdminRepository(db *sqlx.DB) *PostgresAdminRepository {
	return &PostgresAdminRepository{db: db}
}

func (r *PostgresAdminRepository) Count() (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM admin_users`)
	return count, err
}

func (r *PostgresAdminRepository) GetByUsername(username string) (*domain.AdminUser, error) {
	var u domain.AdminUser
	err := r.db.Get(&u, `SELECT * FROM admin_users WHERE username = $1`, username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresAdminRepository) Create(u *domain.AdminUser) error {
	return r.db.QueryRowx(`
		INSERT INTO admin_users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id, created_at`,
		u.Username, u.PasswordHash,
	).Scan(&u.ID, &u.CreatedAt)
}

func (r *PostgresAdminRepository) UpdatePassword(id, passwordHash string) error {
	_, err := r.db.Exec(`UPDATE admin_users SET password_hash = $2 WHERE id = $1`, id, passwordHash)
	return err
}
