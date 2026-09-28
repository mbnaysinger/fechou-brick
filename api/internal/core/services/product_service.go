package services

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/mbnaysinger/fechou-brick-api/internal/core/domain"
)

var (
	ErrInvalidProduct  = errors.New("dados do anúncio inválidos")
	ErrProductNotFound = errors.New("anúncio não encontrado")
	ErrPhotoNotFound   = errors.New("foto não encontrada")
	ErrInvalidPhoto    = errors.New("arquivo de foto inválido")
	ErrInvalidCursor   = errors.New("cursor de paginação inválido")
)

const (
	MaxListLimit     = 50
	DefaultListLimit = 20
)

var allowedPhotoTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

type ProductService struct {
	products domain.ProductRepository
	photos   domain.PhotoRepository
	maxPhotoBytes int
}

func NewProductService(products domain.ProductRepository, photos domain.PhotoRepository, maxPhotoBytes int) *ProductService {
	return &ProductService{products: products, photos: photos, maxPhotoBytes: maxPhotoBytes}
}

// ---------------------------------------------------------------------------
// Público
// ---------------------------------------------------------------------------

type PublicList struct {
	Items      []domain.ProductListItem `json:"items"`
	NextCursor *string                  `json:"next_cursor"`
}

func (s *ProductService) ListPublic(cursor string, limit int) (*PublicList, error) {
	if limit <= 0 {
		limit = DefaultListLimit
	}
	if limit > MaxListLimit {
		limit = MaxListLimit
	}

	afterTitle, afterID := "", ""
	if cursor != "" {
		var err error
		afterTitle, afterID, err = decodeCursor(cursor)
		if err != nil {
			return nil, ErrInvalidCursor
		}
	}

	// Busca limit+1 para saber se existe próxima página sem query extra
	items, err := s.products.ListActive(afterTitle, afterID, limit+1)
	if err != nil {
		return nil, err
	}

	var next *string
	if len(items) > limit {
		items = items[:limit]
		last := items[len(items)-1]
		c := encodeCursor(strings.ToLower(last.Title), last.ID)
		next = &c
	}
	return &PublicList{Items: items, NextCursor: next}, nil
}

type ProductDetail struct {
	domain.Product
	Photos []domain.PhotoMeta `json:"photos"`
}

// GetPublicBySlug retorna ATIVO e VENDIDO (o FE mostra faixa "Vendido"); OCULTO = 404
func (s *ProductService) GetPublicBySlug(slug string) (*ProductDetail, error) {
	p, err := s.products.GetBySlug(slug)
	if err != nil {
		return nil, err
	}
	if p == nil || p.Status == domain.StatusOculto {
		return nil, ErrProductNotFound
	}
	return s.withPhotos(p)
}

func (s *ProductService) withPhotos(p *domain.Product) (*ProductDetail, error) {
	photos, err := s.photos.ListMetaByProduct(p.ID)
	if err != nil {
		return nil, err
	}
	return &ProductDetail{Product: *p, Photos: photos}, nil
}

func (s *ProductService) GetPhoto(id string) (*domain.Photo, error) {
	ph, err := s.photos.GetByID(id)
	if err != nil {
		return nil, err
	}
	if ph == nil {
		return nil, ErrPhotoNotFound
	}
	return ph, nil
}

// ---------------------------------------------------------------------------
// Admin
// ---------------------------------------------------------------------------

type ProductInput struct {
	Title            string   `json:"title"`
	Description      *string  `json:"description"`
	PricePix         float64  `json:"price_pix"`
	PriceCardTotal   *float64 `json:"price_card_total"`
	CardInstallments *int     `json:"card_installments"`
	WhatsappURL      string   `json:"whatsapp_url"`
}

func (in *ProductInput) validate() error {
	in.Title = strings.TrimSpace(in.Title)
	in.WhatsappURL = strings.TrimSpace(in.WhatsappURL)
	if in.Title == "" {
		return fmt.Errorf("%w: título é obrigatório", ErrInvalidProduct)
	}
	if in.PricePix <= 0 {
		return fmt.Errorf("%w: valor no PIX deve ser maior que zero", ErrInvalidProduct)
	}
	if in.WhatsappURL == "" {
		return fmt.Errorf("%w: URL do WhatsApp é obrigatória", ErrInvalidProduct)
	}
	if !strings.HasPrefix(in.WhatsappURL, "https://") {
		return fmt.Errorf("%w: URL do WhatsApp deve começar com https://", ErrInvalidProduct)
	}
	if in.PriceCardTotal == nil {
		// Somente PIX: parcelas não fazem sentido
		in.CardInstallments = nil
		return nil
	}
	if *in.PriceCardTotal <= 0 {
		return fmt.Errorf("%w: valor total no cartão deve ser maior que zero", ErrInvalidProduct)
	}
	if in.CardInstallments != nil && (*in.CardInstallments < 1 || *in.CardInstallments > 12) {
		return fmt.Errorf("%w: parcelas devem estar entre 1 e 12", ErrInvalidProduct)
	}
	return nil
}

func (s *ProductService) ListAdmin() ([]domain.ProductListItem, error) {
	return s.products.ListAll()
}

func (s *ProductService) GetAdminByID(id string) (*ProductDetail, error) {
	p, err := s.products.GetByID(id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProductNotFound
	}
	return s.withPhotos(p)
}

func (s *ProductService) Create(in ProductInput) (*domain.Product, error) {
	if err := in.validate(); err != nil {
		return nil, err
	}
	slug, err := s.uniqueSlug(in.Title)
	if err != nil {
		return nil, err
	}
	p := &domain.Product{
		Slug:             slug,
		Title:            in.Title,
		Description:      in.Description,
		PricePix:         in.PricePix,
		PriceCardTotal:   in.PriceCardTotal,
		CardInstallments: in.CardInstallments,
		WhatsappURL:      in.WhatsappURL,
		Status:           domain.StatusAtivo,
	}
	if err := s.products.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProductService) Update(id string, in ProductInput) (*domain.Product, error) {
	if err := in.validate(); err != nil {
		return nil, err
	}
	p, err := s.products.GetByID(id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProductNotFound
	}
	// Slug é estável: links já compartilhados continuam funcionando mesmo se o título mudar
	p.Title = in.Title
	p.Description = in.Description
	p.PricePix = in.PricePix
	p.PriceCardTotal = in.PriceCardTotal
	p.CardInstallments = in.CardInstallments
	p.WhatsappURL = in.WhatsappURL
	if err := s.products.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProductService) UpdateStatus(id, status string) (*domain.Product, error) {
	if !domain.IsValidStatus(status) {
		return nil, fmt.Errorf("%w: status deve ser ATIVO, VENDIDO ou OCULTO", ErrInvalidProduct)
	}
	p, err := s.products.UpdateStatus(id, status)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProductNotFound
	}
	return p, nil
}

func (s *ProductService) Delete(id string) error {
	p, err := s.products.GetByID(id)
	if err != nil {
		return err
	}
	if p == nil {
		return ErrProductNotFound
	}
	return s.products.Delete(id)
}

type PhotoUpload struct {
	ContentType string
	Data        []byte
}

func (s *ProductService) AddPhotos(productID string, uploads []PhotoUpload) ([]domain.PhotoMeta, error) {
	p, err := s.products.GetByID(productID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProductNotFound
	}
	if len(uploads) == 0 {
		return nil, fmt.Errorf("%w: nenhum arquivo enviado", ErrInvalidPhoto)
	}
	for _, u := range uploads {
		if !allowedPhotoTypes[u.ContentType] {
			return nil, fmt.Errorf("%w: tipo %s não suportado (use JPEG, PNG ou WebP)", ErrInvalidPhoto, u.ContentType)
		}
		if len(u.Data) == 0 || len(u.Data) > s.maxPhotoBytes {
			return nil, fmt.Errorf("%w: cada foto deve ter entre 1 byte e %d MB", ErrInvalidPhoto, s.maxPhotoBytes/(1024*1024))
		}
	}
	pos, err := s.photos.NextPosition(productID)
	if err != nil {
		return nil, err
	}
	for _, u := range uploads {
		ph := &domain.Photo{
			ProductID:   productID,
			Position:    pos,
			ContentType: u.ContentType,
			SizeBytes:   len(u.Data),
			Data:        u.Data,
		}
		if err := s.photos.Add(ph); err != nil {
			return nil, err
		}
		pos++
	}
	return s.photos.ListMetaByProduct(productID)
}

func (s *ProductService) ReorderPhotos(productID string, orderedIDs []string) ([]domain.PhotoMeta, error) {
	p, err := s.products.GetByID(productID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProductNotFound
	}
	if len(orderedIDs) == 0 {
		return nil, fmt.Errorf("%w: lista de ordenação vazia", ErrInvalidPhoto)
	}
	if err := s.photos.Reorder(productID, orderedIDs); err != nil {
		return nil, err
	}
	return s.photos.ListMetaByProduct(productID)
}

func (s *ProductService) DeletePhoto(photoID string) error {
	ph, err := s.photos.GetByID(photoID)
	if err != nil {
		return err
	}
	if ph == nil {
		return ErrPhotoNotFound
	}
	return s.photos.Delete(photoID)
}

// ---------------------------------------------------------------------------
// Slug e cursor
// ---------------------------------------------------------------------------

// Slugify normaliza o título: remove acentos, minúsculas, hífens
func Slugify(title string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	clean, _, err := transform.String(t, title)
	if err != nil {
		clean = title
	}
	var b strings.Builder
	lastHyphen := true // evita hífen no início
	for _, r := range strings.ToLower(clean) {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			lastHyphen = false
		case !lastHyphen:
			b.WriteRune('-')
			lastHyphen = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func (s *ProductService) uniqueSlug(title string) (string, error) {
	base := Slugify(title)
	if base == "" {
		base = "anuncio"
	}
	slug := base
	for i := 2; ; i++ {
		exists, err := s.products.SlugExists(slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}

// Cursor = base64url("lower(title)\x00id") — \x00 nunca aparece em texto
func encodeCursor(lowerTitle, id string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(lowerTitle + "\x00" + id))
}

func decodeCursor(cursor string) (title, id string, err error) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return "", "", err
	}
	parts := strings.SplitN(string(raw), "\x00", 2)
	if len(parts) != 2 || parts[1] == "" {
		return "", "", errors.New("formato inválido")
	}
	return parts[0], parts[1], nil
}
