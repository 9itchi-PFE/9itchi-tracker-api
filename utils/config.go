package utils

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Port        int
	Environment string
	// DatabaseURL is the PostgreSQL connection string (env DATABASE_URL). Never log it.
	DatabaseURL string
	// SecretKey signs JWTs (HS256), EncryptionKey wraps them with AES-GCM. Shared by every service.
	SecretKey     string
	EncryptionKey string
	// InternalAPIKey is the shared secret for service-to-service calls (env INTERNAL_API_KEY).
	InternalAPIKey string
	// InternalAPIKeyRequired (env INTERNAL_API_KEY_REQUIRED, default true); false = log and allow.
	InternalAPIKeyRequired bool
	RabbitMQURL            string
	// Peer service base URLs.
	EmailAPIURL string
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() AppConfig {
	viper.AutomaticEnv()
	return AppConfig{
		Port:                   getIntWithDefault("PORT", 8086),
		Environment:            getStringWithDefault("ENVIRONMENT", "development"),
		DatabaseURL:            getStringWithDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/tracking?sslmode=disable"),
		SecretKey:              getStringWithDefault("ITCHI_SECRET_KEY", ""),
		EncryptionKey:          getStringWithDefault("ITCHI_ENCRYPTION_KEY", ""),
		InternalAPIKey:         getStringWithDefault("INTERNAL_API_KEY", ""),
		InternalAPIKeyRequired: getBoolWithDefault("INTERNAL_API_KEY_REQUIRED", true),
		RabbitMQURL:            getStringWithDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		EmailAPIURL:            getStringWithDefault("EMAIL_API_URL", "http://9itchi-email-api:8085"),
	}
}

// Validate fails fast on an unusable config. Only variable names are reported, never values.
func (c AppConfig) Validate() error {
	if c.Port < 1 {
		return fmt.Errorf("invalid PORT: must be a positive integer")
	}
	var bad []string
	if len(c.SecretKey) < 32 {
		bad = append(bad, "ITCHI_SECRET_KEY (min 32 chars)")
	}
	if len(c.EncryptionKey) < 32 {
		bad = append(bad, "ITCHI_ENCRYPTION_KEY (min 32 chars)")
	}
	if c.DatabaseURL == "" {
		bad = append(bad, "DATABASE_URL")
	}
	if c.InternalAPIKeyRequired && c.InternalAPIKey == "" {
		bad = append(bad, "INTERNAL_API_KEY")
	}
	if len(bad) > 0 {
		return fmt.Errorf("required environment variables are missing or invalid: %s", strings.Join(bad, ", "))
	}
	return nil
}

func getStringWithDefault(key, defaultValue string) string {
	viper.SetDefault(key, defaultValue)
	return viper.GetString(key)
}

func getIntWithDefault(key string, defaultValue int) int {
	viper.SetDefault(key, defaultValue)
	return viper.GetInt(key)
}

func getBoolWithDefault(key string, defaultValue bool) bool {
	viper.SetDefault(key, defaultValue)
	return viper.GetBool(key)
}
