package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/mbnaysinger/fechou-brick-api/internal/core/services"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// writeServiceError mapeia erros de negócio para status HTTP sem vazar detalhes internos
func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidProduct),
		errors.Is(err, services.ErrInvalidPhoto),
		errors.Is(err, services.ErrInvalidCursor),
		errors.Is(err, services.ErrWeakPassword):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, services.ErrProductNotFound), errors.Is(err, services.ErrPhotoNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, services.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, err.Error())
	default:
		slog.Error("erro interno no handler", "erro", err.Error())
		writeError(w, http.StatusInternalServerError, "Erro interno ao processar a requisição")
	}
}

func decodeBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "Corpo da requisição inválido")
		return false
	}
	return true
}
