package services

import (
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/mbnaysinger/fechou-brick-api/internal/core/domain"
)

type fakeAdminRepo struct {
	users map[string]*domain.AdminUser
}

func newFakeAdminRepo() *fakeAdminRepo { return &fakeAdminRepo{users: map[string]*domain.AdminUser{}} }

func (f *fakeAdminRepo) Count() (int, error) { return len(f.users), nil }

func (f *fakeAdminRepo) GetByUsername(username string) (*domain.AdminUser, error) {
	u, ok := f.users[username]
	if !ok {
		return nil, nil
	}
	cp := *u
	return &cp, nil
}

func (f *fakeAdminRepo) Create(u *domain.AdminUser) error {
	u.ID = "admin-1"
	cp := *u
	f.users[u.Username] = &cp
	return nil
}

func (f *fakeAdminRepo) UpdatePassword(id, hash string) error {
	for _, u := range f.users {
		if u.ID == id {
			u.PasswordHash = hash
			return nil
		}
	}
	return errors.New("não existe")
}

func newAuthService() (*AuthService, *fakeAdminRepo) {
	repo := newFakeAdminRepo()
	s := NewAuthService(repo, "segredo-de-teste")
	s.sleep = func(time.Duration) {} // não dormir nos testes
	return s, repo
}

func TestSeedAdminCriaUsuarioUnico(t *testing.T) {
	s, repo := newAuthService()
	if err := s.SeedAdmin("admin", "senha-inicial"); err != nil {
		t.Fatal(err)
	}
	u, _ := repo.GetByUsername("admin")
	if u == nil {
		t.Fatal("admin não criado")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("senha-inicial")) != nil {
		t.Error("hash não confere com a senha inicial")
	}

	// Segunda subida com outra senha não sobrescreve
	if err := s.SeedAdmin("admin", "outra-senha"); err != nil {
		t.Fatal(err)
	}
	u2, _ := repo.GetByUsername("admin")
	if u2.PasswordHash != u.PasswordHash {
		t.Error("seed sobrescreveu usuário existente")
	}
}

func TestSeedAdminSemSenhaFalha(t *testing.T) {
	s, _ := newAuthService()
	if err := s.SeedAdmin("admin", ""); err == nil {
		t.Error("esperava erro com tabela vazia e senha inicial ausente")
	}
}

func TestLogin(t *testing.T) {
	s, _ := newAuthService()
	if err := s.SeedAdmin("admin", "senha-correta"); err != nil {
		t.Fatal(err)
	}

	token, err := s.Login("admin", "senha-correta")
	if err != nil || token == "" {
		t.Fatalf("login válido falhou: %v", err)
	}

	if _, err := s.Login("admin", "senha-errada"); !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("senha errada: esperava ErrInvalidCredentials, veio %v", err)
	}
	if _, err := s.Login("inexistente", "x"); !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("usuário inexistente: esperava ErrInvalidCredentials, veio %v", err)
	}
}

func TestChangePassword(t *testing.T) {
	s, _ := newAuthService()
	if err := s.SeedAdmin("admin", "senha-antiga"); err != nil {
		t.Fatal(err)
	}

	if err := s.ChangePassword("admin", "errada", "nova-senha-longa"); !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("senha atual errada: esperava ErrInvalidCredentials, veio %v", err)
	}
	if err := s.ChangePassword("admin", "senha-antiga", "curta"); !errors.Is(err, ErrWeakPassword) {
		t.Errorf("senha fraca: esperava ErrWeakPassword, veio %v", err)
	}
	if err := s.ChangePassword("admin", "senha-antiga", "nova-senha-longa"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Login("admin", "nova-senha-longa"); err != nil {
		t.Errorf("login com a nova senha falhou: %v", err)
	}
}
