package config

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	strongSecret := strings.Repeat("s", minSecretKeyLength)
	strongPassword := "a-strong-admin-pw"

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{"development allows defaults", Config{AppEnv: "development", SecretKey: defaultSecretKey, AdminPassword: defaultAdminPassword}, false},
		{"production with strong values", Config{AppEnv: "production", SecretKey: strongSecret, AdminPassword: strongPassword}, false},
		{"production with default secret", Config{AppEnv: "production", SecretKey: defaultSecretKey, AdminPassword: strongPassword}, true},
		{"production with short secret", Config{AppEnv: "production", SecretKey: "short", AdminPassword: strongPassword}, true},
		{"production with default admin password", Config{AppEnv: "production", SecretKey: strongSecret, AdminPassword: defaultAdminPassword}, true},
		{"production with short admin password", Config{AppEnv: "production", SecretKey: strongSecret, AdminPassword: "short"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
