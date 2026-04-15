package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration, loaded once at startup.
type Config struct {
	DB     DBConfig
	JWT    JWTConfig
	Server ServerConfig
	Alpha  AlphaConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	SigningKey string
	TokenTTL  time.Duration
}

type ServerConfig struct {
	Host         string
	Port         string
	CookieSecure bool
	CORSOrigins  []string
}

type AlphaConfig struct {
	AlphaVantageKey string
	GoldPricezKey   string
}

// Load reads environment variables (optionally from .env) and returns a validated Config.
func Load() (*Config, error) {
	godotenv.Load() // .env file is optional

	cfg := &Config{}

	// Database
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		if err := cfg.DB.parseURL(databaseURL); err != nil {
			return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
		}
	} else {
		cfg.DB = DBConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		}
	}

	// JWT
	cfg.JWT = JWTConfig{
		SigningKey: os.Getenv("JWT_SIGNING_KEY"),
		TokenTTL:  60 * time.Minute,
	}

	// Server
	cfg.Server = ServerConfig{
		Host:         os.Getenv("HOST"),
		Port:         getEnv("PORT", "8080"),
		CookieSecure: parseBool(os.Getenv("COOKIE_SECURE")),
		CORSOrigins:  parseCORSOrigins(os.Getenv("CORS_ALLOWED_ORIGINS")),
	}

	// External APIs
	cfg.Alpha = AlphaConfig{
		AlphaVantageKey: os.Getenv("ALPHA_VANTAGE_API_KEY"),
		GoldPricezKey:   os.Getenv("GOLD_PRICEZ_API_KEY"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	log.Printf("Config loaded: db=%s:%s/%s server=%s:%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, cfg.Server.Host, cfg.Server.Port)

	return cfg, nil
}

func (db *DBConfig) parseURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if u.Host == "" {
		return fmt.Errorf("DATABASE_URL has no host")
	}
	parts := strings.Split(u.Host, ":")
	db.Host = parts[0]
	if len(parts) > 1 {
		db.Port = parts[1]
	} else {
		db.Port = "5432"
	}
	db.User = u.User.Username()
	db.Password, _ = u.User.Password()
	db.Name = strings.TrimPrefix(u.Path, "/")
	db.SSLMode = u.Query().Get("sslmode")
	if db.SSLMode == "" {
		db.SSLMode = "disable"
	}
	return nil
}

func (c *Config) validate() error {
	if c.DB.Host == "" {
		return fmt.Errorf("DB_HOST is required (set DATABASE_URL or DB_HOST)")
	}
	if c.JWT.SigningKey == "" {
		return fmt.Errorf("JWT_SIGNING_KEY is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func parseBool(s string) bool {
	v := strings.ToLower(strings.TrimSpace(s))
	return v == "1" || v == "true" || v == "yes"
}

func parseCORSOrigins(s string) []string {
	configured := strings.TrimSpace(s)
	if configured == "" {
		return []string{
			"http://localhost:3100",
			"http://127.0.0.1:3100",
			"https://primetrading-nine.vercel.app",
		}
	}
	parts := strings.Split(configured, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		origin := strings.TrimSpace(p)
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	if len(origins) == 0 {
		return []string{"http://localhost:3100", "http://127.0.0.1:3100"}
	}
	return origins
}
