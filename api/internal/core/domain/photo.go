package domain

import "time"

type Photo struct {
	ID          string    `db:"id"`
	ProductID   string    `db:"product_id"`
	Position    int       `db:"position"`
	ContentType string    `db:"content_type"`
	SizeBytes   int       `db:"size_bytes"`
	Data        []byte    `db:"data"`
	CreatedAt   time.Time `db:"created_at"`
}

// PhotoMeta trafega em listagens e no detalhe do produto (sem os bytes)
type PhotoMeta struct {
	ID          string    `db:"id" json:"id"`
	ProductID   string    `db:"product_id" json:"product_id"`
	Position    int       `db:"position" json:"position"`
	ContentType string    `db:"content_type" json:"content_type"`
	SizeBytes   int       `db:"size_bytes" json:"size_bytes"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
