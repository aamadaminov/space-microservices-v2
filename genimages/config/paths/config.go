package paths

import (
	"os"
	"path/filepath"
)

type Config struct {
	ImgPath       string
	DirPathSource string
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Load() (Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	return Config{
		ImgPath:       getEnv("IMG_PATH", filepath.Join(homeDir, "images")),
		DirPathSource: getEnv("DIRSOURCE_PATH", "./sourceimages"),
	}, nil
}
