package kafka

import "os"

type Config struct {
	AddressKafka string
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Load() (Config, error) {
	return Config{
		AddressKafka: getEnv("ADDRESS_KAFKA", "172.17.0.1:9092"),
	}, nil
}
