package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set required environment variables
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg == nil {
		t.Fatal("Config is nil")
	}

	// Verify default values
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected server port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Database.Name != "airygithub" {
		t.Errorf("Expected database name 'airygithub', got %s", cfg.Database.Name)
	}

	if cfg.JWT.Secret != "test-secret-key" {
		t.Errorf("Expected JWT secret 'test-secret-key', got %s", cfg.JWT.Secret)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Server:   ServerConfig{Port: 8080},
				JWT:      JWTConfig{Secret: "valid-secret"},
				Database: DatabaseConfig{Name: "test"},
				Pool:     PoolConfig{Size: 100},
				Security: SecurityConfig{BcryptCost: 10},
			},
			wantErr: false,
		},
		{
			name: "invalid port",
			config: &Config{
				Server:   ServerConfig{Port: 0},
				JWT:      JWTConfig{Secret: "valid-secret"},
				Database: DatabaseConfig{Name: "test"},
				Pool:     PoolConfig{Size: 100},
				Security: SecurityConfig{BcryptCost: 10},
			},
			wantErr: true,
		},
		{
			name: "empty JWT secret",
			config: &Config{
				Server:   ServerConfig{Port: 8080},
				JWT:      JWTConfig{Secret: ""},
				Database: DatabaseConfig{Name: "test"},
				Pool:     PoolConfig{Size: 100},
				Security: SecurityConfig{BcryptCost: 10},
			},
			wantErr: true,
		},
		{
			name: "default JWT secret",
			config: &Config{
				Server:   ServerConfig{Port: 8080},
				JWT:      JWTConfig{Secret: "change-me-in-production"},
				Database: DatabaseConfig{Name: "test"},
				Pool:     PoolConfig{Size: 100},
				Security: SecurityConfig{BcryptCost: 10},
			},
			wantErr: true,
		},
		{
			name: "empty database name",
			config: &Config{
				Server:   ServerConfig{Port: 8080},
				JWT:      JWTConfig{Secret: "valid-secret"},
				Database: DatabaseConfig{Name: ""},
				Pool:     PoolConfig{Size: 100},
				Security: SecurityConfig{BcryptCost: 10},
			},
			wantErr: true,
		},
		{
			name: "invalid pool size",
			config: &Config{
				Server:   ServerConfig{Port: 8080},
				JWT:      JWTConfig{Secret: "valid-secret"},
				Database: DatabaseConfig{Name: "test"},
				Pool:     PoolConfig{Size: 0},
				Security: SecurityConfig{BcryptCost: 10},
			},
			wantErr: true,
		},
		{
			name: "invalid bcrypt cost too low",
			config: &Config{
				Server:   ServerConfig{Port: 8080},
				JWT:      JWTConfig{Secret: "valid-secret"},
				Database: DatabaseConfig{Name: "test"},
				Pool:     PoolConfig{Size: 100},
				Security: SecurityConfig{BcryptCost: 3},
			},
			wantErr: true,
		},
		{
			name: "invalid bcrypt cost too high",
			config: &Config{
				Server:   ServerConfig{Port: 8080},
				JWT:      JWTConfig{Secret: "valid-secret"},
				Database: DatabaseConfig{Name: "test"},
				Pool:     PoolConfig{Size: 100},
				Security: SecurityConfig{BcryptCost: 32},
			},
			wantErr: true,
		},
		{
			name: "TLS enabled without cert file",
			config: &Config{
				Server:   ServerConfig{Port: 8080, TLSEnabled: true, TLSKeyFile: "/path/to/key.pem"},
				JWT:      JWTConfig{Secret: "valid-secret"},
				Database: DatabaseConfig{Name: "test"},
				Pool:     PoolConfig{Size: 100},
				Security: SecurityConfig{BcryptCost: 10},
			},
			wantErr: true,
		},
		{
			name: "TLS enabled without key file",
			config: &Config{
				Server:   ServerConfig{Port: 8080, TLSEnabled: true, TLSCertFile: "/path/to/cert.pem"},
				JWT:      JWTConfig{Secret: "valid-secret"},
				Database: DatabaseConfig{Name: "test"},
				Pool:     PoolConfig{Size: 100},
				Security: SecurityConfig{BcryptCost: 10},
			},
			wantErr: true,
		},
		{
			name: "TLS enabled with both files",
			config: &Config{
				Server:   ServerConfig{Port: 8080, TLSEnabled: true, TLSCertFile: "/path/to/cert.pem", TLSKeyFile: "/path/to/key.pem"},
				JWT:      JWTConfig{Secret: "valid-secret"},
				Database: DatabaseConfig{Name: "test"},
				Pool:     PoolConfig{Size: 100},
				Security: SecurityConfig{BcryptCost: 10},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetDSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		Name:     "testdb",
		User:     "testuser",
		Password: "testpass",
	}

	expected := "testuser:testpass@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	got := cfg.GetDSN()

	if got != expected {
		t.Errorf("GetDSN() = %s, want %s", got, expected)
	}
}

func TestGetRedisAddr(t *testing.T) {
	cfg := RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	expected := "localhost:6379"
	got := cfg.GetAddr()

	if got != expected {
		t.Errorf("GetAddr() = %s, want %s", got, expected)
	}
}

func TestGetESAddr(t *testing.T) {
	cfg := ESConfig{
		Host: "localhost",
		Port: 9200,
	}

	expected := "http://localhost:9200"
	got := cfg.GetAddr()

	if got != expected {
		t.Errorf("GetAddr() = %s, want %s", got, expected)
	}
}

func TestGetMQAddr(t *testing.T) {
	cfg := MQConfig{
		Host:     "localhost",
		Port:     5672,
		User:     "guest",
		Password: "guest",
	}

	expected := "amqp://guest:guest@localhost:5672/"
	got := cfg.GetAddr()

	if got != expected {
		t.Errorf("GetAddr() = %s, want %s", got, expected)
	}
}
