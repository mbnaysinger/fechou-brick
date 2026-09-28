package repositories

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/mbnaysinger/fechou-brick-api/internal/core/domain"
)

type PostgresPhotoRepository struct {
	db *sqlx.DB
}

func NewPostgresPhotoRepository(db *sqlx.DB) *PostgresPhotoRepository {
	return &PostgresPhotoRepository{db: db}
}

func (r *PostgresPhotoRepository) GetByID(id string) (*domain.Photo, error) {
	var p domain.Photo
	err := r.db.Get(&p, `SELECT * FROM product_photos WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostgresPhotoRepository) ListMetaByProduct(productID string) ([]domain.PhotoMeta, error) {
	metas := []domain.PhotoMeta{}
	err := r.db.Select(&metas, `
		SELECT id, product_id, position, content_type, size_bytes, created_at
		FROM product_photos
		WHERE product_id = $1
		ORDER BY position, created_at`, productID)
	return metas, err
}

func (r *PostgresPhotoRepository) Add(p *domain.Photo) error {
	return r.db.QueryRowx(`
		INSERT INTO product_photos (product_id, position, content_type, size_bytes, data)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`,
		p.ProductID, p.Position, p.ContentType, p.SizeBytes, p.Data,
	).Scan(&p.ID, &p.CreatedAt)
}

func (r *PostgresPhotoRepository) NextPosition(productID string) (int, error) {
	var next int
	err := r.db.Get(&next, `
		SELECT COALESCE(MAX(position) + 1, 0) FROM product_photos WHERE product_id = $1`, productID)
	return next, err
}

func (r *PostgresPhotoRepository) Reorder(productID string, orderedIDs []string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range orderedIDs {
		if _, err := tx.Exec(`
			UPDATE product_photos SET position = $3
			WHERE id = $1 AND product_id = $2`, id, productID, i); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *PostgresPhotoRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM product_photos WHERE id = $1`, id)
	return err
}
