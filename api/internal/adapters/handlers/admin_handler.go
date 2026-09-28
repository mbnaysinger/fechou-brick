package handlers

import (
	"io"
	"net/http"

	"github.com/mbnaysinger/fechou-brick-api/internal/adapters/middlewares"
	"github.com/mbnaysinger/fechou-brick-api/internal/core/services"
)

type AdminHandler struct {
	products      *services.ProductService
	auth          *services.AuthService
	maxPhotoBytes int64
}

func NewAdminHandler(products *services.ProductService, auth *services.AuthService, maxPhotoBytes int64) *AdminHandler {
	return &AdminHandler{products: products, auth: auth, maxPhotoBytes: maxPhotoBytes}
}

// ---------------------------------------------------------------------------
// Auth
// ---------------------------------------------------------------------------

func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	token, err := h.auth.Login(body.Username, body.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *AdminHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	username, ok := middlewares.UsernameFromRequest(r)
	if !ok {
		writeError(w, http.StatusInternalServerError, "Falha ao ler o contexto da requisição")
		return
	}
	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	if err := h.auth.ChangePassword(username, body.CurrentPassword, body.NewPassword); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Senha atualizada"})
}

// ---------------------------------------------------------------------------
// Produtos
// ---------------------------------------------------------------------------

func (h *AdminHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.products.ListAdmin()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *AdminHandler) Get(w http.ResponseWriter, r *http.Request) {
	detail, err := h.products.GetAdminByID(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (h *AdminHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in services.ProductInput
	if !decodeBody(w, r, &in) {
		return
	}
	p, err := h.products.Create(in)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *AdminHandler) Update(w http.ResponseWriter, r *http.Request) {
	var in services.ProductInput
	if !decodeBody(w, r, &in) {
		return
	}
	p, err := h.products.Update(r.PathValue("id"), in)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *AdminHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	p, err := h.products.UpdateStatus(r.PathValue("id"), body.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *AdminHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.products.Delete(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Fotos
// ---------------------------------------------------------------------------

// UploadPhotos recebe multipart no campo "fotos" (1..n arquivos).
// O Content-Type real é detectado dos bytes, não confiado no header do cliente.
func (h *AdminHandler) UploadPhotos(w http.ResponseWriter, r *http.Request) {
	// Teto do corpo inteiro: 6 fotos no limite individual, folga suficiente
	r.Body = http.MaxBytesReader(w, r.Body, h.maxPhotoBytes*6)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "Upload inválido ou acima do limite de tamanho")
		return
	}
	defer r.MultipartForm.RemoveAll()

	files := r.MultipartForm.File["fotos"]
	uploads := make([]services.PhotoUpload, 0, len(files))
	for _, fh := range files {
		f, err := fh.Open()
		if err != nil {
			writeError(w, http.StatusBadRequest, "Falha ao ler arquivo enviado")
			return
		}
		data, err := io.ReadAll(io.LimitReader(f, h.maxPhotoBytes+1))
		f.Close()
		if err != nil {
			writeError(w, http.StatusBadRequest, "Falha ao ler arquivo enviado")
			return
		}
		contentType := http.DetectContentType(firstBytes(data))
		uploads = append(uploads, services.PhotoUpload{ContentType: contentType, Data: data})
	}

	metas, err := h.products.AddPhotos(r.PathValue("id"), uploads)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, metas)
}

func firstBytes(data []byte) []byte {
	if len(data) > 512 {
		return data[:512]
	}
	return data
}

func (h *AdminHandler) ReorderPhotos(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OrderedIDs []string `json:"ordered_ids"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	metas, err := h.products.ReorderPhotos(r.PathValue("id"), body.OrderedIDs)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, metas)
}

func (h *AdminHandler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	if err := h.products.DeletePhoto(r.PathValue("foto_id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
