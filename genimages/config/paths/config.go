package paths

import "os"

type Config struct {
    ImgPath         string
    DirPathSource   string
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func Load() (Config, error) {
    return Config{
        ImgPath:            getEnv("IMG_PATH", "./images/"),
        DirPathSource:      getEnv("DIRSOURCE_PATH", "./sourceimages/"),
    }, nil
}
