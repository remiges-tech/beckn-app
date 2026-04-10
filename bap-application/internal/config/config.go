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

	// BPP identity — used inside the Beckn context fields (bppId / bppUri).
	BppID  string // BPP subscriber ID  (context.bppId)
	BppURI string // BPP receiver URL   (context.bppUri, e.g. http://bpptest.remiges.tech/bpp/receiver)

	// AdapterURL is the ONIX BAP adapter base URL the BAP actually POSTs to.
	// The action is appended: <AdapterURL>/select, <AdapterURL>/init, etc.
	// This is intentionally different from BppURI.
	AdapterURL string

	// CDSDiscoverURL is the full URL of the Catalog Discovery Service discover endpoint.
	// Example: https://cds.example.com/discover
	// Leave blank in environments without a live CDS (returns empty catalog list).
	CDSDiscoverURL string
}

// Load reads all config values from the environment and returns a validated Config.
func Load() (*Config, error) {
	cfg := &Config{
		AppPort: getEnv("APP_PORT", "8083"),
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

		BppID:      os.Getenv("BPP_ID"),
		BppURI:     os.Getenv("BPP_URI"),
		AdapterURL: os.Getenv("ADAPTER_URL"),

		CDSDiscoverURL: os.Getenv("CDS_DISCOVER_URL"),
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
	if c.BppID == "" {
		return fmt.Errorf("BPP_ID is required")
	}
	if c.BppURI == "" {
		return fmt.Errorf("BPP_URI is required")
	}
	if c.AdapterURL == "" {
		return fmt.Errorf("ADAPTER_URL is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
