package domain

type ProductRepository interface {
	// ListActive pagina por keyset: (lower(title), id) > (afterTitle, afterID).
	// afterID vazio = primeira página.
	ListActive(afterTitle, afterID string, limit int) ([]ProductListItem, error)
	ListAll() ([]ProductListItem, error)
	GetBySlug(slug string) (*Product, error)
	GetByID(id string) (*Product, error)
	SlugExists(slug string) (bool, error)
	Create(p *Product) error
	Update(p *Product) error
	UpdateStatus(id, status string) (*Product, error)
	Delete(id string) error
}

type PhotoRepository interface {
	GetByID(id string) (*Photo, error)
	ListMetaByProduct(productID string) ([]PhotoMeta, error)
	Add(p *Photo) error
	NextPosition(productID string) (int, error)
	// Reorder aplica position = índice na lista; ids fora da lista mantêm posição
	Reorder(productID string, orderedIDs []string) error
	Delete(id string) error
}

type AdminUserRepository interface {
	Count() (int, error)
	GetByUsername(username string) (*AdminUser, error)
	Create(u *AdminUser) error
	UpdatePassword(id, passwordHash string) error
}
