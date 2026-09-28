package domain

import "time"

const (
	StatusAtivo   = "ATIVO"
	StatusVendido = "VENDIDO"
	StatusOculto  = "OCULTO"
)

func IsValidStatus(s string) bool {
	return s == StatusAtivo || s == StatusVendido || s == StatusOculto
}

type Product struct {
	ID               string    `db:"id" json:"id"`
	Slug             string    `db:"slug" json:"slug"`
	Title            string    `db:"title" json:"title"`
	Description      *string   `db:"description" json:"description"`
	PricePix         float64   `db:"price_pix" json:"price_pix"`
	PriceCardTotal   *float64  `db:"price_card_total" json:"price_card_total"`
	CardInstallments *int      `db:"card_installments" json:"card_installments"`
	WhatsappURL      string    `db:"whatsapp_url" json:"whatsapp_url"`
	Status           string    `db:"status" json:"status"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// ProductListItem é a projeção da listagem: nunca carrega bytes de foto,
// só o id da principal para o FE montar a URL.
type ProductListItem struct {
	Product
	PrimaryPhotoID *string `db:"primary_photo_id" json:"primary_photo_id"`
}
