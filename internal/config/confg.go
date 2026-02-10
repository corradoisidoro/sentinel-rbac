package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort  int
	DatabaseURL string
	JWTSecret   string
}

func Load() (Config, error) {
	// Best-effort .env loading (dev only)
	_ = godotenv.Load()

	cfg := Config{
		ServerPort: 3000, // safe default
	}

	if v := os.Getenv("SERVER_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("config: invalid SERVER_PORT: %w", err)
		}
		cfg.ServerPort = p
	}

	if v := os.Getenv("DATABASE_URL"); v != "" {
		cfg.DatabaseURL = v
	}

	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWTSecret = v
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if c.ServerPort <= 0 || c.ServerPort > 65535 {
		return errors.New("config: invalid SERVER_PORT")
	}
	if c.DatabaseURL == "" {
		return errors.New("config: missing DATABASE_URL")
	}
	if c.JWTSecret == "" {
		return errors.New("config: missing JWT_SECRET")
	}
	return nil
}
