package handlers

import (
	"net/http"
	"strconv"

	"github.com/mbnaysinger/fechou-brick-api/internal/core/services"
)

type PhotoHandler struct {
	service *services.ProductService
}

func NewPhotoHandler(service *services.ProductService) *PhotoHandler {
	return &PhotoHandler{service: service}
}

// Serve responde os bytes da foto com cache imutável: o id nunca muda de
// conteúdo, então o navegador baixa cada foto uma única vez.
func (h *PhotoHandler) Serve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	etag := `"` + id + `"`

	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	photo, err := h.service.GetPhoto(id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", photo.ContentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(photo.Data)))
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("ETag", etag)
	w.WriteHeader(http.StatusOK)
	w.Write(photo.Data)
}
