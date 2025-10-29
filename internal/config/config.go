// ==========================
// internal/config/config.go (UPDATED)
// ==========================
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// M-Pesa API Credentials
	ConsumerKey    string
	ConsumerSecret string

	// Business Configuration
	BusinessShortCode int
	Passkey           string
	InitiatorName     string
	InitiatorPassword string
	CertificatePath   string // Path to M-Pesa public certificate

	// API URLs
	BaseURL string

	// Callback URLs
	CallbackBaseURL string
	STKCallbackURL  string
	B2CResultURL    string
	B2CTimeoutURL   string
	B2BResultURL    string
	B2BTimeoutURL   string

	// Server Configuration
	Host   string
	Port   int
	Debug  bool
	Reload bool

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// Logging
	LogLevel string

	// API Timeout
	APITimeout int
}

func (c *Config) OAuthURL() string {
	return fmt.Sprintf("%s/oauth/v1/generate?grant_type=client_credentials", c.BaseURL)
}

func (c *Config) STKPushURL() string {
	return fmt.Sprintf("%s/mpesa/stkpush/v1/processrequest", c.BaseURL)
}

func (c *Config) STKQueryURL() string {
	return fmt.Sprintf("%s/mpesa/stkpushquery/v1/query", c.BaseURL)
}

func (c *Config) B2CURL() string {
	return fmt.Sprintf("%s/mpesa/b2c/v3/paymentrequest", c.BaseURL)
}

func Load() (*Config, error) {
	// Load .env file
	_ = godotenv.Load()

	shortCode, err := strconv.Atoi(getEnv("BUSINESS_SHORT_CODE", ""))
	if err != nil {
		return nil, fmt.Errorf("invalid BUSINESS_SHORT_CODE: %w", err)
	}

	port, _ := strconv.Atoi(getEnv("PORT", "8000"))
	apiTimeout, _ := strconv.Atoi(getEnv("API_TIMEOUT", "30"))

	return &Config{
		ConsumerKey:       getEnv("CONSUMER_KEY", ""),
		ConsumerSecret:    getEnv("CONSUMER_SECRET", ""),
		BusinessShortCode: shortCode,
		Passkey:           getEnv("PASSKEY", ""),
		InitiatorName:     getEnv("INITIATOR_NAME", ""),
		InitiatorPassword: getEnv("INITIATOR_PASSWORD", ""),
		CertificatePath:   getEnv("CERTIFICATE_PATH", "certs/SandboxCertificate.cer"),
		BaseURL:           getEnv("BASE_URL", ""),
		CallbackBaseURL:   getEnv("CALLBACK_BASE_URL", ""),
		STKCallbackURL:    getEnv("STK_CALLBACK_URL", ""),
		B2CResultURL:      getEnv("B2C_RESULT_URL", ""),
		B2CTimeoutURL:     getEnv("B2C_TIMEOUT_URL", ""),
		B2BResultURL:      getEnv("B2B_RESULT_URL", ""),
		B2BTimeoutURL:     getEnv("B2B_TIMEOUT_URL", ""),
		Host:              getEnv("HOST", "0.0.0.0"),
		Port:              port,
		Debug:             getEnv("DEBUG", "true") == "true",
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		RedisURL:          getEnv("REDIS_URL", ""),
		LogLevel:          getEnv("LOG_LEVEL", "INFO"),
		APITimeout:        apiTimeout,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
