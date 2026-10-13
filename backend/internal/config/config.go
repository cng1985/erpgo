package config

import (
	"os"
	"strings"
	"time"
	_ "time/tzdata"
)

type Config struct {
	Address        string
	DatabasePath   string
	AllowedOrigins []string
	TimeZone       string
}

func Load() (Config, error) {
	timeZone := env("APP_TIMEZONE", "Asia/Shanghai")
	location, err := time.LoadLocation(timeZone)
	if err != nil {
		return Config{}, err
	}
	time.Local = location
	origins := strings.Split(env("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173"), ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}
	return Config{
		Address:        env("APP_ADDR", ":8080"),
		DatabasePath:   env("DB_PATH", "data/erp.db"),
		AllowedOrigins: origins,
		TimeZone:       timeZone,
	}, nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
