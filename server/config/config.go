package config

import (
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort     string
	SecretKey   string
	DatabaseURL string
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
		AppPort:     strOrDefault("APP_PORT", "8080"),
		SecretKey:   secretKey,
		DatabaseURL: dbURL,
	}, nil
}

func strOrDefault(key, def string) string {
	if v := viper.GetString(key); v != "" {
		return v
	}
	return def
}
