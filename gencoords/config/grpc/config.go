package grpc

import "os"

type Config struct {
    AddressGrpc    string
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func Load() (Config, error) {
    return Config{
        AddressGrpc:    getEnv("ADDRESS_GRPC", ":50070"),
    }, nil
}
