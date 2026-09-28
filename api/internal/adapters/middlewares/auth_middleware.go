package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const usernameKey contextKey = "username"

type AuthMiddleware struct {
	jwtSecret []byte
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: []byte(jwtSecret)}
}

// RequireAuth valida o Bearer JWT (HS256) emitido pelo login do admin
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		tokenStr, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok || tokenStr == "" {
			unauthorized(w)
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return m.jwtSecret, nil
		}, jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !token.Valid {
			unauthorized(w)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			unauthorized(w)
			return
		}
		sub, _ := claims["sub"].(string)
		if sub == "" {
			unauthorized(w)
			return
		}

		ctx := context.WithValue(r.Context(), usernameKey, sub)
		next(w, r.WithContext(ctx))
	}
}

func UsernameFromRequest(r *http.Request) (string, bool) {
	username, ok := r.Context().Value(usernameKey).(string)
	return username, ok
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error": "Não autorizado"}`))
}
