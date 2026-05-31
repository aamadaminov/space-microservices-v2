package otel

import "os"

type Config struct {
    OtelExporterEndpoint    string
    OtelServiceName string
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func Load() (Config, error) {
    return Config{
        OtelExporterEndpoint:    getEnv("OTEL_ENDPOINT", "127.0.0.1:4317"),
        OtelServiceName: "CoordsGen",
    }, nil
}
