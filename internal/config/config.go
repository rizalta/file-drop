package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort       string
	StoragePath      string
	DBUser           string
	DBPassword       string
	DBHost           string
	DBPort           string
	DBName           string
	MaxUploadSize    int64
	RateLimitUploads int
}

func Load() (*Config, error) {
	maxUploadSizeStr := getEnv("MAX_UPLOAD_SIZE", "100")
	maxUploadSize, err := strconv.ParseInt(maxUploadSizeStr, 10, 64)
	if err != nil {
		return nil, err
	}

	rateLimitUploadsStr := getEnv("RATE_LIMIT_UPLOADS", "100")
	rateLimitUploads, err := strconv.Atoi(rateLimitUploadsStr)
	if err != nil {
		return nil, err
	}

	return &Config{
		ServerPort:       getEnv("PORT", "8000"),
		StoragePath:      getEnv("STORAGE_PATH", "./blobs"),
		DBUser:           getEnv("DB_USER", "postgres"),
		DBPassword:       getEnv("DB_PASSWORD", "postgres"),
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBName:           getEnv("DB_NAME", "file_drop"),
		MaxUploadSize:    maxUploadSize,
		RateLimitUploads: rateLimitUploads,
	}, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
