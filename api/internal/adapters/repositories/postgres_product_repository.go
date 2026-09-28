package repositories

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/mbnaysinger/fechou-brick-api/internal/core/domain"
)

type PostgresProductRepository struct {
	db *sqlx.DB
}

func NewPostgresProductRepository(db *sqlx.DB) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

// primaryPhotoSelect resolve o id da foto principal sem carregar bytes
const primaryPhotoSelect = `(
	SELECT ph.id FROM product_photos ph
	WHERE ph.product_id = p.id
	ORDER BY ph.position, ph.created_at
	LIMIT 1
) AS primary_photo_id`

func (r *PostgresProductRepository) ListActive(afterTitle, afterID string, limit int) ([]domain.ProductListItem, error) {
	items := []domain.ProductListItem{}
	if afterID == "" {
		err := r.db.Select(&items, `
			SELECT p.*, `+primaryPhotoSelect+`
			FROM products p
			WHERE p.status = 'ATIVO'
			ORDER BY lower(p.title), p.id
			LIMIT $1`, limit)
		return items, err
	}
	err := r.db.Select(&items, `
		SELECT p.*, `+primaryPhotoSelect+`
		FROM products p
		WHERE p.status = 'ATIVO'
		  AND (lower(p.title), p.id) > ($1, $2::uuid)
		ORDER BY lower(p.title), p.id
		LIMIT $3`, afterTitle, afterID, limit)
	return items, err
}

func (r *PostgresProductRepository) ListAll() ([]domain.ProductListItem, error) {
	items := []domain.ProductListItem{}
	err := r.db.Select(&items, `
		SELECT p.*, `+primaryPhotoSelect+`
		FROM products p
		ORDER BY p.created_at DESC`)
	return items, err
}

func (r *PostgresProductRepository) GetBySlug(slug string) (*domain.Product, error) {
	return r.getOne(`SELECT * FROM products WHERE slug = $1`, slug)
}

func (r *PostgresProductRepository) GetByID(id string) (*domain.Product, error) {
	return r.getOne(`SELECT * FROM products WHERE id = $1`, id)
}

func (r *PostgresProductRepository) getOne(query string, arg any) (*domain.Product, error) {
	var p domain.Product
	err := r.db.Get(&p, query, arg)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostgresProductRepository) SlugExists(slug string) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, `SELECT EXISTS (SELECT 1 FROM products WHERE slug = $1)`, slug)
	return exists, err
}

func (r *PostgresProductRepository) Create(p *domain.Product) error {
	return r.db.QueryRowx(`
		INSERT INTO products (slug, title, description, price_pix, price_card_total, card_installments, whatsapp_url, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`,
		p.Slug, p.Title, p.Description, p.PricePix, p.PriceCardTotal, p.CardInstallments, p.WhatsappURL, p.Status,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *PostgresProductRepository) Update(p *domain.Product) error {
	return r.db.QueryRowx(`
		UPDATE products
		SET title = $2, description = $3, price_pix = $4, price_card_total = $5,
		    card_installments = $6, whatsapp_url = $7, updated_at = now()
		WHERE id = $1
		RETURNING updated_at`,
		p.ID, p.Title, p.Description, p.PricePix, p.PriceCardTotal, p.CardInstallments, p.WhatsappURL,
	).Scan(&p.UpdatedAt)
}

func (r *PostgresProductRepository) UpdateStatus(id, status string) (*domain.Product, error) {
	var p domain.Product
	err := r.db.Get(&p, `
		UPDATE products SET status = $2, updated_at = now()
		WHERE id = $1
		RETURNING *`, id, status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostgresProductRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM products WHERE id = $1`, id)
	return err
}
