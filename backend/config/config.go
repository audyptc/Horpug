package config

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv                 string
	AppPort                string
	DBHost                 string
	DBPort                 string
	DBUsername             string
	DBPassword             string
	DBName                 string
	DBSSLMode              string
	SecretKey              string
	AccessTokenTTL         time.Duration
	RefreshTokenTTL        time.Duration
	AdminUsername          string
	AdminEmail             string
	AdminPassword          string
	CookieSecure           bool
	LineChannelAccessToken string
	LineChannelID          string
	UploadDir              string
	PublicURL              string
}

func Load() Config {
	_ = godotenv.Load(".env")

	return Config{
		// AppEnv is "production" on the deployed server (set by
		// .github/workflows/deploy.yml); Validate only enforces its checks there
		// so local dev keeps working with the defaults below.
		AppEnv:          getEnv("APP_ENV", "development"),
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

		// LineChannelAccessToken is the Messaging API channel's token, used
		// to push messages through the dormitory's LINE OA.
		//
		// LineChannelID is NOT that same channel — LINE no longer allows
		// attaching a LIFF app to a Messaging API channel, so the LIFF app
		// (frontend/src/features/line/LineLinkPage.tsx) lives under a
		// separate LINE Login channel created in the same Provider. This
		// must be that LINE Login channel's Channel ID (Basic settings tab),
		// since it's used as the client_id when verifying the LIFF id token
		// (its "aud" claim is the channel that owns the LIFF app).
		LineChannelAccessToken: getEnv("LINE_CHANNEL_ACCESS_TOKEN", ""),
		LineChannelID:          getEnv("LINE_CHANNEL_ID", ""),

		// UploadDir is where uploaded files (e.g. documents) are stored. In
		// Docker it is a mounted volume so files survive container rebuilds.
		UploadDir: getEnv("UPLOAD_DIR", "./uploads"),

		// PublicURL is the site's public address, e.g. https://horpug.example.com.
		// LINE downloads images from public HTTPS URLs only, so the PromptPay
		// QR image in overdue reminders is sent only when this is https://.
		PublicURL: getEnv("APP_PUBLIC_URL", ""),
	}
}

const (
	defaultSecretKey     = "change-me-in-production"
	defaultAdminPassword = "Admin@12345"
	minSecretKeyLength   = 32
	minAdminPasswordLen  = 12
)

func (c Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// Validate refuses to start production with a missing or default secret: the
// JWT key and the admin password fall back to well-known values, which would
// let anyone forge tokens or log in as admin. SeedAdmin re-applies
// ADMIN_PASSWORD on every start, so a default here resets the live admin.
func (c Config) Validate() error {
	if !c.IsProduction() {
		return nil
	}

	var errs []error
	if c.SecretKey == defaultSecretKey || len(c.SecretKey) < minSecretKeyLength {
		errs = append(errs, errors.New("APP_SECRETKEY must be set to a random value of at least 32 characters"))
	}
	if c.AdminPassword == defaultAdminPassword || len(c.AdminPassword) < minAdminPasswordLen {
		errs = append(errs, errors.New("ADMIN_PASSWORD must be set to a non-default value of at least 12 characters"))
	}
	return errors.Join(errs...)
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
