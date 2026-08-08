package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rizalta/file-drop/db"
	"github.com/rizalta/file-drop/internal/config"
	"github.com/rizalta/file-drop/internal/handler"
	"github.com/rizalta/file-drop/internal/repo"
	"github.com/rizalta/file-drop/internal/service"
	"github.com/rizalta/file-drop/internal/storage"
	"github.com/rizalta/file-drop/web"
)

func main() {
	ctx := context.Background()
	config := config.Load()

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.DBUser,
		config.DBPassword,
		config.DBHost,
		config.DBPort,
		config.DBName,
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}

	if _, err := pool.Exec(ctx, db.SchemaSQL); err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}

	r := chi.NewRouter()
	spaHandler, err := web.SPAHandler()
	if err != nil {
		log.Fatalf("failed to initialize static handler")
	}

	r.Use(middleware.Logger)

	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := pool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status": "error",
				"db":     "disconnected",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"db":     "connected",
		})
	})

	querier := repo.New(pool)
	storage, err := storage.NewFileStorage(config.StoragePath)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}

	service := service.NewService(querier, storage)

	handler := handler.NewHandler(service)

	r.Post("/api/upload", handler.UploadDrop)
	r.Get("/api/f/{id}", handler.GetDrop)
	r.Delete("/api/f/{id}", handler.DeleteDrop)

	r.Handle("/*", spaHandler)

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if err := service.CleanupExpiredDrops(ctx); err != nil {
				log.Printf("failed to cleanup drops: %v", err)
			}
		}
	}()

	fmt.Printf("Server is running on port: %s\n", config.ServerPort)
	if err := http.ListenAndServe(":"+config.ServerPort, r); err != nil {
		log.Fatalf("failed to Listen to %s: %v", config.ServerPort, err)
	}
}
