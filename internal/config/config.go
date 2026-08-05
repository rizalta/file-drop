package config

import "os"

type Config struct {
	ServerPort  string
	StoragePath string
	DBUser      string
	DBPassword  string
	DBHost      string
	DBPort      string
	DBName      string
}

func Load() *Config {
	return &Config{
		ServerPort:  getEnv("PORT", "8080"),
		StoragePath: getEnv("STORAGE_PATH", "./blobs"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "postgres"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBName:      getEnv("DB_NAME", "file_drop"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
