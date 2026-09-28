package services

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/mbnaysinger/fechou-brick-api/internal/core/domain"
)

var (
	ErrInvalidCredentials = errors.New("usuário ou senha inválidos")
	ErrWeakPassword       = errors.New("a nova senha deve ter pelo menos 8 caracteres")
)

const tokenTTL = 24 * time.Hour

// failedLoginDelay dificulta força bruta sem precisar de rate-limit elaborado
const failedLoginDelay = 300 * time.Millisecond

type AuthService struct {
	users     domain.AdminUserRepository
	jwtSecret []byte
	sleep     func(time.Duration) // injetável para os testes não esperarem de verdade
}

func NewAuthService(users domain.AdminUserRepository, jwtSecret string) *AuthService {
	return &AuthService{users: users, jwtSecret: []byte(jwtSecret), sleep: time.Sleep}
}

// SeedAdmin cria o usuário único na primeira subida, se a tabela estiver vazia
func (s *AuthService) SeedAdmin(username, initialPassword string) error {
	count, err := s.users.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if initialPassword == "" {
		return errors.New("nenhum admin cadastrado e ADMIN_INITIAL_PASSWORD não definida")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(initialPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.users.Create(&domain.AdminUser{Username: username, PasswordHash: string(hash)}); err != nil {
		return err
	}
	slog.Info("Usuário admin criado (seed inicial)", slog.String("username", username))
	return nil
}

func (s *AuthService) Login(username, password string) (string, error) {
	user, err := s.users.GetByUsername(username)
	if err != nil {
		return "", err
	}
	if user == nil {
		s.sleep(failedLoginDelay)
		return "", ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.sleep(failedLoginDelay)
		return "", ErrInvalidCredentials
	}

	claims := jwt.MapClaims{
		"sub": user.Username,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(tokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("falha ao assinar token: %w", err)
	}
	return signed, nil
}

func (s *AuthService) ChangePassword(username, currentPassword, newPassword string) error {
	user, err := s.users.GetByUsername(username)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		s.sleep(failedLoginDelay)
		return ErrInvalidCredentials
	}
	if len(newPassword) < 8 {
		return ErrWeakPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(user.ID, string(hash))
}
