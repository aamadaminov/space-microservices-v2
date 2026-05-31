package metrics

import "os"

type Config struct {
    AddressMetrics    string
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func Load() (Config, error) {
    return Config{
        AddressMetrics:    getEnv("ADDRESS_METRICS", ":2223"),
    }, nil
}
