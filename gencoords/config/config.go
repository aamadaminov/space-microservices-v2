package config

import (
    "github.com/aamadaminov/space-microservices-v2/gencoords/config/metrics"
    "github.com/aamadaminov/space-microservices-v2/gencoords/config/otel"
)

type Config struct {
    OTEL    otel.Config
    Metrics metrics.Config
}

func Init() (Config, error) {
    var cfg Config
    var err error
    
    // Загружаем конфиги из подпакетов
    cfg.OTEL, err = otel.Load()
    if err != nil {
        return Config{}, err
    }
    
    cfg.Metrics, err = metrics.Load()
    if err != nil {
        return Config{}, err
    }
    
    return cfg, nil
}
