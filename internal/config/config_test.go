package config_test

import (
	"testing"

	"github.com/corradoisidoro/sentinel-rbac/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLoad_Success(t *testing.T) {
	t.Setenv("SERVER_PORT", "8080")
	t.Setenv("DATABASE_URL", "test.db")
	t.Setenv("JWT_SECRET", "super-secret")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.Equal(t, 8080, cfg.ServerPort)
	require.Equal(t, "test.db", cfg.DatabaseURL)
	require.Equal(t, "super-secret", cfg.JWTSecret)
}

func TestLoad_DefaultPort(t *testing.T) {
	t.Setenv("DATABASE_URL", "test.db")
	t.Setenv("JWT_SECRET", "secret")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.Equal(t, 3000, cfg.ServerPort)
}

func TestLoad_InvalidServerPort(t *testing.T) {
	tests := []string{
		"abc",
		"-1",
		"0",
		"70000",
	}

	for _, v := range tests {
		t.Run(v, func(t *testing.T) {
			t.Setenv("SERVER_PORT", v)
			t.Setenv("DATABASE_URL", "test.db")
			t.Setenv("JWT_SECRET", "secret")

			_, err := config.Load()
			require.Error(t, err)
		})
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	t.Setenv("JWT_SECRET", "secret")

	_, err := config.Load()

	require.Error(t, err)
	require.Contains(t, err.Error(), "DATABASE_URL")
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "test.db")

	_, err := config.Load()

	require.Error(t, err)
	require.Contains(t, err.Error(), "JWT_SECRET")
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name  string
		cfg   config.Config
		valid bool
	}{
		{
			name: "valid config",
			cfg: config.Config{
				ServerPort:  3000,
				DatabaseURL: "db",
				JWTSecret:   "secret",
			},
			valid: true,
		},
		{
			name: "invalid port",
			cfg: config.Config{
				ServerPort:  -1,
				DatabaseURL: "db",
				JWTSecret:   "secret",
			},
			valid: false,
		},
		{
			name: "missing database url",
			cfg: config.Config{
				ServerPort: 3000,
				JWTSecret:  "secret",
			},
			valid: false,
		},
		{
			name: "missing jwt secret",
			cfg: config.Config{
				ServerPort:  3000,
				DatabaseURL: "db",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
