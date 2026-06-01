package config

import (
    "github.com/aamadaminov/space-microservices-v2/genimages/config/paths"
)

type Config struct {
    Paths   paths.Config
}

func Init() (Config, error) {
    var cfg Config
    var err error
    
    // Загружаем конфиги из подпакетов
    cfg.Paths, err = paths.Load()
    if err != nil {
        return Config{}, err
    }
    
    return cfg, nil
}
