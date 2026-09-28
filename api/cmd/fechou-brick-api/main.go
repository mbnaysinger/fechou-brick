package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	"github.com/mbnaysinger/fechou-brick-api/internal/adapters/handlers"
	"github.com/mbnaysinger/fechou-brick-api/internal/adapters/middlewares"
	"github.com/mbnaysinger/fechou-brick-api/internal/adapters/repositories"
	"github.com/mbnaysinger/fechou-brick-api/internal/config"
	"github.com/mbnaysinger/fechou-brick-api/internal/core/services"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
	slog.Info("Iniciando serviço fechou-brick-api")

	godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	if dsn == "" {
		slog.Error("Variável DATABASE_URL ausente")
		os.Exit(1)
	}
	if jwtSecret == "" {
		slog.Error("Variável JWT_SECRET ausente")
		os.Exit(1)
	}

	maxUploadMB := 10
	if v := os.Getenv("MAX_UPLOAD_MB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxUploadMB = n
		}
	}
	maxPhotoBytes := maxUploadMB * 1024 * 1024

	db := config.ConnectDB(dsn)
	defer db.Close()
	runMigrations(db)

	// Repositories
	productRepo := repositories.NewPostgresProductRepository(db)
	photoRepo := repositories.NewPostgresPhotoRepository(db)
	adminRepo := repositories.NewPostgresAdminRepository(db)

	// Services
	productService := services.NewProductService(productRepo, photoRepo, maxPhotoBytes)
	authService := services.NewAuthService(adminRepo, jwtSecret)

	// Seed do usuário único na primeira subida
	adminUsername := os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		adminUsername = "admin"
	}
	if err := authService.SeedAdmin(adminUsername, os.Getenv("ADMIN_INITIAL_PASSWORD")); err != nil {
		slog.Error("Falha no seed do admin", "erro", err.Error())
		os.Exit(1)
	}

	// Handlers
	productHandler := handlers.NewProductHandler(productService)
	photoHandler := handlers.NewPhotoHandler(productService)
	adminHandler := handlers.NewAdminHandler(productService, authService, int64(maxPhotoBytes))

	authMiddleware := middlewares.NewAuthMiddleware(jwtSecret)
	mux := handlers.SetupRoutes(productHandler, photoHandler, adminHandler, authMiddleware)

	loggedRouter := middlewares.LoggerMiddleware(mux)
	corsRouter := middlewares.CorsMiddleware(loggedRouter)

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8085"
	}
	slog.Info("Servidor escutando chamadas", slog.String("port", httpPort))
	if err := http.ListenAndServe(":"+httpPort, corsRouter); err != nil {
		slog.Error("Falha crítica no servidor Web", "erro", err.Error())
		os.Exit(1)
	}
}

func runMigrations(db *sqlx.DB) {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		slog.Error("Não instanciou driver migração", "erro", err.Error())
		os.Exit(1)
	}
	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)
	if err != nil {
		slog.Error("Não inicializou migração", "erro", err.Error())
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("Erro crítico nas migrações", "erro", err.Error())
		os.Exit(1)
	}
	slog.Info("Migrações de banco de dados atualizadas com sucesso!")
}
