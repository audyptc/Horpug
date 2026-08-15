package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort         string
	DBHost          string
	DBPort          string
	DBUsername      string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	SecretKey       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	AdminUsername   string
	AdminEmail      string
	AdminPassword   string
	CookieSecure    bool
}

func Load() Config {
	_ = godotenv.Load(".env")

	return Config{
		AppPort:         getEnv("APP_PORT", "8080"),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBUsername:      getEnv("DB_USERNAME", "postgres"),
		DBPassword:      getEnv("DB_PASSWORD", "password"),
		DBName:          getEnv("DB_NAME", "goapi_db"),
		DBSSLMode:       getEnv("DB_SSLMODE", "disable"),
		SecretKey:       getEnv("APP_SECRETKEY", "change-me-in-production"),
		AccessTokenTTL:  getEnvDuration("APP_ACCESS_TOKEN_TTL", 8*time.Hour),
		RefreshTokenTTL: getEnvDuration("APP_REFRESH_TOKEN_TTL", 7*24*time.Hour),
		AdminUsername:   getEnv("ADMIN_USERNAME", "admin"),
		AdminEmail:      getEnv("ADMIN_EMAIL", "admin@horpug.local"),
		AdminPassword:   getEnv("ADMIN_PASSWORD", "Admin@12345"),
		// The refresh-token cookie's Secure flag. Defaults to false so bare
		// `go run` / `npm run dev` local dev (plain HTTP) still works; set to
		// true in backend/.env when running behind the dockerized nginx, which
		// now terminates HTTPS (see nginx/nginx.conf), or browsers will
		// silently drop the cookie.
		CookieSecure: getEnvBool("APP_COOKIE_SECURE", false),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return fallback
}
