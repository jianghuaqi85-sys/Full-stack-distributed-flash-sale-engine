package config

import (
	"os"
	"testing"
)

func TestLoadDefaultConfig(t *testing.T) {
	// Set required env vars
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing")
	defer os.Unsetenv("JWT_SECRET")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.AppEnv != "production" {
		t.Fatalf("Expected AppEnv 'production', got '%s'", cfg.AppEnv)
	}

	if cfg.AppPort != 8080 {
		t.Fatalf("Expected AppPort 8080, got %d", cfg.AppPort)
	}

	if cfg.LogLevel != "info" {
		t.Fatalf("Expected LogLevel 'info', got '%s'", cfg.LogLevel)
	}

	if cfg.Database.MaxOpenConns != 50 {
		t.Fatalf("Expected MaxOpenConns 50, got %d", cfg.Database.MaxOpenConns)
	}

	if cfg.Database.MaxIdleConns != 25 {
		t.Fatalf("Expected MaxIdleConns 25, got %d", cfg.Database.MaxIdleConns)
	}

	if cfg.Database.ConnMaxIdleTime != 60 {
		t.Fatalf("Expected ConnMaxIdleTime 60, got %d", cfg.Database.ConnMaxIdleTime)
	}

	if cfg.Kafka.Enabled {
		t.Fatal("Expected Kafka disabled by default")
	}

	if cfg.WsGateway.Enabled {
		t.Fatal("Expected WsGateway disabled by default")
	}
}

func TestValidateProduction(t *testing.T) {
	cfg := &Config{
		AppEnv: "production",
		JWT: JWTConfig{
			Secret: "default_secret_key_must_be_overridden_in_production",
		},
		Database: DatabaseConfig{
			MaxOpenConns:    50,
			MaxIdleConns:    25,
			ConnMaxLifetime: 300,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() should fail with default JWT secret in production")
	}

	cfg.JWT.Secret = "strong-secret-key"
	err = cfg.Validate()
	if err != nil {
		t.Fatalf("Validate() should pass with valid config: %v", err)
	}
}

func TestDSN(t *testing.T) {
	db := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		Name:     "testdb",
		SSLMode:  "disable",
	}

	dsn := db.DSN()
	expected := "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable"
	if dsn != expected {
		t.Fatalf("DSN() = '%s', want '%s'", dsn, expected)
	}
}

func TestRedisAddr(t *testing.T) {
	redis := RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	addr := redis.Addr()
	if addr != "localhost:6379" {
		t.Fatalf("Addr() = '%s', want 'localhost:6379'", addr)
	}
}

func TestParseOrigins(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"http://localhost:3000,http://localhost:5173", []string{"http://localhost:3000", "http://localhost:5173"}},
		{"http://example.com", []string{"http://example.com"}},
		{"", []string{}},
		{" , , ", []string{}},
	}

	for _, tt := range tests {
		result := parseOrigins(tt.input)
		if len(result) != len(tt.expected) {
			t.Fatalf("parseOrigins(%q) returned %d items, want %d", tt.input, len(result), len(tt.expected))
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Fatalf("parseOrigins(%q)[%d] = %q, want %q", tt.input, i, v, tt.expected[i])
			}
		}
	}
}
