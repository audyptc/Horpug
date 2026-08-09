package config

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort              string
	SecretKey            string
	DatabaseURL          string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	CORSOrigins          []string
	UploadDir            string
	UploadBaseURL        string
}

func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			log.Println("⚠️  ไม่พบไฟล์ .env ระบบจะใช้ Environment Variables จาก OS")
		} else {
			log.Printf("⚠️  โหลดไฟล์ config ไม่สำเร็จ: %v", err)
		}
	}

	secretKey := viper.GetString("APP_SECRETKEY")
	if secretKey == "" {
		return nil, fmt.Errorf("APP_SECRETKEY ต้องระบุใน .env")
	}

	accessTokenDuration, err := durationOrDefault("ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return nil, err
	}

	refreshTokenDuration, err := durationOrDefault("REFRESH_TOKEN_TTL", 7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		strOrDefault("DB_USERNAME", "postgres"),
		url.QueryEscape(strOrDefault("DB_PASSWORD", "password")),
		strOrDefault("DB_HOST", "localhost"),
		strOrDefault("DB_PORT", "5432"),
		strOrDefault("DB_NAME", "goapi_db"),
		strOrDefault("DB_SSLMODE", "disable"),
	)

	return &Config{
		AppPort:              strOrDefault("APP_PORT", "8080"),
		SecretKey:            secretKey,
		DatabaseURL:          dbURL,
		AccessTokenDuration:  accessTokenDuration,
		RefreshTokenDuration: refreshTokenDuration,
		CORSOrigins:          corsOrigins("CORS_ORIGINS", "http://localhost,http://localhost:5173,http://localhost:3000"),
		UploadDir:            strOrDefault("UPLOAD_DIR", "./uploads"),
		UploadBaseURL:        strOrDefault("UPLOAD_BASE_URL", "http://localhost:8080"),
	}, nil
}

func corsOrigins(key, def string) []string {
	v := viper.GetString(key)
	if v == "" {
		v = def
	}
	var result []string
	for _, s := range strings.Split(v, ",") {
		if s = strings.TrimSpace(s); s != "" {
			result = append(result, s)
		}
	}
	return result
}

func strOrDefault(key, def string) string {
	if v := viper.GetString(key); v != "" {
		return v
	}
	return def
}

func durationOrDefault(key string, def time.Duration) (time.Duration, error) {
	v := viper.GetString(key)
	if v == "" {
		return def, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s ต้องเป็น duration ที่ถูกต้อง เช่น 15m, 1h, 168h", key)
	}
	if d <= 0 {
		return 0, fmt.Errorf("%s ต้องมีค่ามากกว่า 0", key)
	}
	return d, nil
}
