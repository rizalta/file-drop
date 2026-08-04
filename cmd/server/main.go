package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rizalta/file-drop/db"
	"github.com/rizalta/file-drop/internal/handler"
	"github.com/rizalta/file-drop/internal/repo"
	"github.com/rizalta/file-drop/internal/service"
	"github.com/rizalta/file-drop/internal/storage"
	"github.com/rizalta/file-drop/web"
)

func main() {
	ctx := context.Background()

	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	storagePath := os.Getenv("STORAGE_PATH")
	if storagePath == "" {
		storagePath = "./blobs"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("error connecting to db: %v", err)
	}

	if _, err := pool.Exec(ctx, db.SchemaSQL); err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}

	r := chi.NewRouter()

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
	storage, err := storage.NewFileStorage(storagePath)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}

	service := service.NewService(querier, storage)

	handler := handler.NewHandler(service)

	r.Post("/api/upload", handler.UploadDrop)
	r.Get("/f/{id}", handler.DownloadDrop)
	r.Delete("/api/files/{id}", handler.DeleteDrop)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, web.WebUI, "index.html")
	})

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if err := service.CleanupExpiredDrops(ctx); err != nil {
				log.Printf("failed to cleanup drops: %v", err)
			}
		}
	}()

	fmt.Printf("Server is running on port: %s\n", serverPort)
	if err := http.ListenAndServe(":"+serverPort, r); err != nil {
		log.Fatalf("failed to Listen to %s: %v", serverPort, err)
	}
}
