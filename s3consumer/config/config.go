package config

import (
	"github.com/aamadaminov/space-microservices-v2/s3consumer/config/minio"
	"github.com/aamadaminov/space-microservices-v2/s3consumer/config/paths"
)

type Config struct {
	Paths paths.Config
	Minio minio.Config
}

func Init() (Config, error) {
	var cfg Config
	var err error

	// Загружаем конфиги из подпакетов
	cfg.Paths, err = paths.Load()
	if err != nil {
		return Config{}, err
	}

	cfg.Minio, err = minio.Load()
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}
