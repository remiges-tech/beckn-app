package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all runtime configuration for the BAP application.
type Config struct {
	AppPort string
	AppEnv  string

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	LogLevel string

	// Beckn network identity for this BAP.
	BapID     string // Subscriber ID registered on the Beckn network
	BapURI    string // Callback URL the BAP exposes (where BPP sends on_* callbacks)
	NetworkID string // Beckn network identifier
}

// Load reads all config values from the environment and returns a validated Config.
func Load() (*Config, error) {
	cfg := &Config{
		AppPort: getEnv("APP_PORT", "8082"),
		AppEnv:  getEnv("APP_ENV", "development"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "bap"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		LogLevel: getEnv("LOG_LEVEL", "info"),

		BapID:     os.Getenv("BAP_ID"),
		BapURI:    os.Getenv("BAP_URI"),
		NetworkID: os.Getenv("NETWORK_ID"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// DatabaseURL builds a pgx-compatible DSN from the config fields.
func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBName, c.DBUser, c.DBPassword, c.DBSSLMode,
	)
}

// DBPortInt returns the DB port as an integer.
func (c *Config) DBPortInt() int {
	p, _ := strconv.Atoi(c.DBPort)
	return p
}

// IsProduction reports whether the application is running in production mode.
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

func (c *Config) validate() error {
	if c.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.DBUser == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.BapID == "" {
		return fmt.Errorf("BAP_ID is required")
	}
	if c.BapURI == "" {
		return fmt.Errorf("BAP_URI is required")
	}
	if c.NetworkID == "" {
		return fmt.Errorf("NETWORK_ID is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
