package services

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/mbnaysinger/fechou-brick-api/internal/core/domain"
)

// ---------------------------------------------------------------------------
// Fakes em memória
// ---------------------------------------------------------------------------

type fakeProductRepo struct {
	products map[string]*domain.Product
	seq      int
}

func newFakeProductRepo() *fakeProductRepo {
	return &fakeProductRepo{products: map[string]*domain.Product{}}
}

func (f *fakeProductRepo) sortedActive() []domain.ProductListItem {
	var items []domain.ProductListItem
	for _, p := range f.products {
		if p.Status == domain.StatusAtivo {
			items = append(items, domain.ProductListItem{Product: *p})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		ti, tj := strings.ToLower(items[i].Title), strings.ToLower(items[j].Title)
		if ti != tj {
			return ti < tj
		}
		return items[i].ID < items[j].ID
	})
	return items
}

func (f *fakeProductRepo) ListActive(afterTitle, afterID string, limit int) ([]domain.ProductListItem, error) {
	all := f.sortedActive()
	var out []domain.ProductListItem
	for _, it := range all {
		if afterID != "" {
			lt := strings.ToLower(it.Title)
			if lt < afterTitle || (lt == afterTitle && it.ID <= afterID) {
				continue
			}
		}
		out = append(out, it)
		if len(out) == limit {
			break
		}
	}
	return out, nil
}

func (f *fakeProductRepo) ListAll() ([]domain.ProductListItem, error) {
	var items []domain.ProductListItem
	for _, p := range f.products {
		items = append(items, domain.ProductListItem{Product: *p})
	}
	return items, nil
}

func (f *fakeProductRepo) GetBySlug(slug string) (*domain.Product, error) {
	for _, p := range f.products {
		if p.Slug == slug {
			cp := *p
			return &cp, nil
		}
	}
	return nil, nil
}

func (f *fakeProductRepo) GetByID(id string) (*domain.Product, error) {
	p, ok := f.products[id]
	if !ok {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

func (f *fakeProductRepo) SlugExists(slug string) (bool, error) {
	p, _ := f.GetBySlug(slug)
	return p != nil, nil
}

func (f *fakeProductRepo) Create(p *domain.Product) error {
	f.seq++
	p.ID = fmt.Sprintf("id-%03d", f.seq)
	cp := *p
	f.products[p.ID] = &cp
	return nil
}

func (f *fakeProductRepo) Update(p *domain.Product) error {
	if _, ok := f.products[p.ID]; !ok {
		return errors.New("não existe")
	}
	cp := *p
	f.products[p.ID] = &cp
	return nil
}

func (f *fakeProductRepo) UpdateStatus(id, status string) (*domain.Product, error) {
	p, ok := f.products[id]
	if !ok {
		return nil, nil
	}
	p.Status = status
	cp := *p
	return &cp, nil
}

func (f *fakeProductRepo) Delete(id string) error {
	delete(f.products, id)
	return nil
}

type fakePhotoRepo struct {
	photos map[string]*domain.Photo
	seq    int
}

func newFakePhotoRepo() *fakePhotoRepo { return &fakePhotoRepo{photos: map[string]*domain.Photo{}} }

func (f *fakePhotoRepo) GetByID(id string) (*domain.Photo, error) {
	p, ok := f.photos[id]
	if !ok {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

func (f *fakePhotoRepo) ListMetaByProduct(productID string) ([]domain.PhotoMeta, error) {
	var metas []domain.PhotoMeta
	for _, p := range f.photos {
		if p.ProductID == productID {
			metas = append(metas, domain.PhotoMeta{ID: p.ID, ProductID: p.ProductID, Position: p.Position, ContentType: p.ContentType, SizeBytes: p.SizeBytes})
		}
	}
	sort.Slice(metas, func(i, j int) bool { return metas[i].Position < metas[j].Position })
	return metas, nil
}

func (f *fakePhotoRepo) Add(p *domain.Photo) error {
	f.seq++
	p.ID = fmt.Sprintf("ph-%03d", f.seq)
	cp := *p
	f.photos[p.ID] = &cp
	return nil
}

func (f *fakePhotoRepo) NextPosition(productID string) (int, error) {
	next := 0
	for _, p := range f.photos {
		if p.ProductID == productID && p.Position >= next {
			next = p.Position + 1
		}
	}
	return next, nil
}

func (f *fakePhotoRepo) Reorder(productID string, orderedIDs []string) error {
	for i, id := range orderedIDs {
		if p, ok := f.photos[id]; ok && p.ProductID == productID {
			p.Position = i
		}
	}
	return nil
}

func (f *fakePhotoRepo) Delete(id string) error {
	delete(f.photos, id)
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const maxTestPhotoBytes = 10 * 1024 * 1024

func newService() (*ProductService, *fakeProductRepo, *fakePhotoRepo) {
	pr, fr := newFakeProductRepo(), newFakePhotoRepo()
	return NewProductService(pr, fr, maxTestPhotoBytes), pr, fr
}

func validInput(title string) ProductInput {
	return ProductInput{Title: title, PricePix: 100, WhatsappURL: "https://wa.me/5551999999999"}
}

func ptrF(v float64) *float64 { return &v }
func ptrI(v int) *int         { return &v }

// ---------------------------------------------------------------------------
// Testes
// ---------------------------------------------------------------------------

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Bicicleta Aro 29":       "bicicleta-aro-29",
		"Fogão 4 bocas — Cônsul": "fogao-4-bocas-consul",
		"  TV   Samsung 50\" ":   "tv-samsung-50",
		"!!!":                    "",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, esperava %q", in, got, want)
		}
	}
}

func TestCreateGeneratesUniqueSlug(t *testing.T) {
	s, _, _ := newService()
	p1, err := s.Create(validInput("Bicicleta Aro 29"))
	if err != nil {
		t.Fatal(err)
	}
	p2, err := s.Create(validInput("Bicicleta Aro 29"))
	if err != nil {
		t.Fatal(err)
	}
	if p1.Slug != "bicicleta-aro-29" || p2.Slug != "bicicleta-aro-29-2" {
		t.Errorf("slugs = %q, %q", p1.Slug, p2.Slug)
	}
}

func TestCreateValidation(t *testing.T) {
	s, _, _ := newService()

	cases := []struct {
		name string
		in   ProductInput
	}{
		{"sem título", ProductInput{PricePix: 10, WhatsappURL: "https://wa.me/1"}},
		{"pix zero", ProductInput{Title: "x", PricePix: 0, WhatsappURL: "https://wa.me/1"}},
		{"sem whatsapp", ProductInput{Title: "x", PricePix: 10}},
		{"whatsapp sem https", ProductInput{Title: "x", PricePix: 10, WhatsappURL: "wa.me/1"}},
		{"cartão zero", func() ProductInput { in := validInput("x"); in.PriceCardTotal = ptrF(0); return in }()},
		{"13 parcelas", func() ProductInput {
			in := validInput("x")
			in.PriceCardTotal = ptrF(100)
			in.CardInstallments = ptrI(13)
			return in
		}()},
	}
	for _, c := range cases {
		if _, err := s.Create(c.in); !errors.Is(err, ErrInvalidProduct) {
			t.Errorf("%s: esperava ErrInvalidProduct, veio %v", c.name, err)
		}
	}
}

func TestCreateSomentePixDescartaParcelas(t *testing.T) {
	s, _, _ := newService()
	in := validInput("Mesa")
	in.CardInstallments = ptrI(4) // parcelas sem valor de cartão: incoerente, deve ser descartado
	p, err := s.Create(in)
	if err != nil {
		t.Fatal(err)
	}
	if p.CardInstallments != nil {
		t.Error("parcelas deveriam ser descartadas quando não há valor de cartão")
	}
}

func TestListPublicPaginationKeyset(t *testing.T) {
	s, repo, _ := newService()
	for _, title := range []string{"Cadeira", "Abajur", "Bicicleta", "Escrivaninha", "Ventilador"} {
		if _, err := s.Create(validInput(title)); err != nil {
			t.Fatal(err)
		}
	}
	// Um VENDIDO não aparece na listagem pública
	sold, _ := s.Create(validInput("Aparador"))
	if _, err := s.UpdateStatus(sold.ID, domain.StatusVendido); err != nil {
		t.Fatal(err)
	}

	page1, err := s.ListPublic("", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page1.Items) != 2 || page1.Items[0].Title != "Abajur" || page1.Items[1].Title != "Bicicleta" {
		t.Fatalf("página 1 inesperada: %+v", page1.Items)
	}
	if page1.NextCursor == nil {
		t.Fatal("esperava next_cursor na página 1")
	}

	page2, err := s.ListPublic(*page1.NextCursor, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page2.Items) != 2 || page2.Items[0].Title != "Cadeira" || page2.Items[1].Title != "Escrivaninha" {
		t.Fatalf("página 2 inesperada: %+v", page2.Items)
	}

	page3, err := s.ListPublic(*page2.NextCursor, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(page3.Items) != 1 || page3.Items[0].Title != "Ventilador" || page3.NextCursor != nil {
		t.Fatalf("página 3 inesperada: %+v (cursor %v)", page3.Items, page3.NextCursor)
	}

	// Estabilidade: item removido entre páginas não quebra o cursor
	all := repo.sortedActive()
	_ = repo.Delete(all[0].ID)
	again, err := s.ListPublic(*page1.NextCursor, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(again.Items) != 2 || again.Items[0].Title != "Cadeira" {
		t.Fatalf("cursor instável após remoção: %+v", again.Items)
	}
}

func TestListPublicCursorInvalido(t *testing.T) {
	s, _, _ := newService()
	if _, err := s.ListPublic("###nao-e-base64###", 10); !errors.Is(err, ErrInvalidCursor) {
		t.Errorf("esperava ErrInvalidCursor, veio %v", err)
	}
}

func TestGetPublicBySlugOcultoEVendido(t *testing.T) {
	s, _, _ := newService()
	p, _ := s.Create(validInput("Guarda Roupa"))

	if _, err := s.UpdateStatus(p.ID, domain.StatusOculto); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetPublicBySlug(p.Slug); !errors.Is(err, ErrProductNotFound) {
		t.Errorf("OCULTO deveria dar not found, veio %v", err)
	}

	if _, err := s.UpdateStatus(p.ID, domain.StatusVendido); err != nil {
		t.Fatal(err)
	}
	det, err := s.GetPublicBySlug(p.Slug)
	if err != nil {
		t.Fatalf("VENDIDO deveria ser acessível pelo link direto: %v", err)
	}
	if det.Status != domain.StatusVendido {
		t.Errorf("status = %s", det.Status)
	}
}

func TestAddPhotosValidaTipoETamanho(t *testing.T) {
	s, _, _ := newService()
	p, _ := s.Create(validInput("Sofá"))

	if _, err := s.AddPhotos(p.ID, []PhotoUpload{{ContentType: "application/pdf", Data: []byte{1}}}); !errors.Is(err, ErrInvalidPhoto) {
		t.Errorf("tipo inválido: esperava ErrInvalidPhoto, veio %v", err)
	}
	big := make([]byte, maxTestPhotoBytes+1)
	if _, err := s.AddPhotos(p.ID, []PhotoUpload{{ContentType: "image/jpeg", Data: big}}); !errors.Is(err, ErrInvalidPhoto) {
		t.Errorf("acima do teto: esperava ErrInvalidPhoto, veio %v", err)
	}

	metas, err := s.AddPhotos(p.ID, []PhotoUpload{
		{ContentType: "image/jpeg", Data: []byte{1, 2}},
		{ContentType: "image/webp", Data: []byte{3}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(metas) != 2 || metas[0].Position != 0 || metas[1].Position != 1 {
		t.Fatalf("metas inesperadas: %+v", metas)
	}
}

func TestReorderPhotosDefineNovaPrincipal(t *testing.T) {
	s, _, _ := newService()
	p, _ := s.Create(validInput("Estante"))
	metas, _ := s.AddPhotos(p.ID, []PhotoUpload{
		{ContentType: "image/jpeg", Data: []byte{1}},
		{ContentType: "image/jpeg", Data: []byte{2}},
	})

	reordered, err := s.ReorderPhotos(p.ID, []string{metas[1].ID, metas[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	if reordered[0].ID != metas[1].ID || reordered[0].Position != 0 {
		t.Fatalf("reordenação falhou: %+v", reordered)
	}
}
