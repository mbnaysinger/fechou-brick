package handlers

import (
	"net/http"

	"github.com/mbnaysinger/fechou-brick-api/internal/adapters/middlewares"
)

func SetupRoutes(
	productHandler *ProductHandler,
	photoHandler *PhotoHandler,
	adminHandler *AdminHandler,
	authMiddleware *middlewares.AuthMiddleware,
) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "UP", "message": "Serviço fechou-brick-api operando normalmente"}`))
	})

	// ---- Público (catálogo) ----
	mux.HandleFunc("GET /api/v1/produtos", productHandler.List)
	mux.HandleFunc("GET /api/v1/produtos/{slug}", productHandler.GetBySlug)
	mux.HandleFunc("GET /api/v1/fotos/{id}", photoHandler.Serve)

	// ---- Admin (usuário único, JWT) ----
	mux.HandleFunc("POST /api/v1/admin/login", adminHandler.Login)

	requireAuth := authMiddleware.RequireAuth
	mux.HandleFunc("PATCH /api/v1/admin/senha", requireAuth(adminHandler.ChangePassword))

	mux.HandleFunc("GET /api/v1/admin/produtos", requireAuth(adminHandler.List))
	mux.HandleFunc("POST /api/v1/admin/produtos", requireAuth(adminHandler.Create))
	mux.HandleFunc("GET /api/v1/admin/produtos/{id}", requireAuth(adminHandler.Get))
	mux.HandleFunc("PUT /api/v1/admin/produtos/{id}", requireAuth(adminHandler.Update))
	mux.HandleFunc("PATCH /api/v1/admin/produtos/{id}/status", requireAuth(adminHandler.UpdateStatus))
	mux.HandleFunc("DELETE /api/v1/admin/produtos/{id}", requireAuth(adminHandler.Delete))

	mux.HandleFunc("POST /api/v1/admin/produtos/{id}/fotos", requireAuth(adminHandler.UploadPhotos))
	mux.HandleFunc("PUT /api/v1/admin/produtos/{id}/fotos/ordem", requireAuth(adminHandler.ReorderPhotos))
	mux.HandleFunc("DELETE /api/v1/admin/fotos/{foto_id}", requireAuth(adminHandler.DeletePhoto))

	return mux
}
