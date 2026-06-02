package minio

import (
	"os"
)

type Config struct {
	MinioAddr     string
	MinioUser     string
	MinioPassword string
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Load() (Config, error) {
	return Config{
		MinioAddr:     getEnv("MINIO_ENDPOINT", "localhost:9050"),
		MinioUser:     getEnv("MINIO_USER", "minioadmin"),
		MinioPassword: getEnv("MINIO_PASSWORD", "minioadmin"),
	}, nil
}
