package config

import (
    "github.com/aamadaminov/space-microservices-v2/gentemps/config/metrics"
    "github.com/aamadaminov/space-microservices-v2/gentemps/config/otel"
    "github.com/aamadaminov/space-microservices-v2/gentemps/config/grpc"
)

type Config struct {
    OTEL    otel.Config
    Metrics metrics.Config
    GRPC    grpc.Config
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

    cfg.GRPC, err = grpc.Load()
    if err != nil {
        return Config{}, err
    }
    
    return cfg, nil
}
