package handlers

import (
	"net/http"
	"strconv"

	"github.com/mbnaysinger/fechou-brick-api/internal/core/services"
)

// ProductHandler cobre as rotas públicas do catálogo
type ProductHandler struct {
	service *services.ProductService
}

func NewProductHandler(service *services.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	cursor := r.URL.Query().Get("cursor")

	list, err := h.service.ListPublic(cursor, limit)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *ProductHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	detail, err := h.service.GetPublicBySlug(r.PathValue("slug"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}
