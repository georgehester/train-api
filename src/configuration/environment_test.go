package configuration

import (
	"testing"
)

func TestLoadPort(t *testing.T) {
	t.Setenv("PORT", "")
	if got := loadPort(); got != "8000" {
		t.Fatalf("expected default port 8000, got %q", got)
	}

	t.Setenv("PORT", "9090")
	if got := loadPort(); got != "9090" {
		t.Fatalf("expected port 9090, got %q", got)
	}
}

func TestLoadApplicationEnvironment(t *testing.T) {
	t.Setenv("ENVIRONMENT", "development")
	if got := loadApplicationEnvironment(); got != EnvironmentDevelopment {
		t.Fatalf("expected development environment, got %q", got)
	}

	t.Setenv("ENVIRONMENT", "production")
	if got := loadApplicationEnvironment(); got != EnvironmentProduction {
		t.Fatalf("expected production environment for non-development value, got %q", got)
	}
}

func TestLoadDatabaseConnectionString(t *testing.T) {
	t.Setenv("DATABASE_USERNAME", "user")
	t.Setenv("DATABASE_PASSWORD", "pass")
	t.Setenv("DATABASE_HOST", "localhost")
	t.Setenv("DATABASE_PORT", "5432")
	t.Setenv("DATABASE_NAME", "trains")

	connectionString := loadDatabaseConnectionString()
	expected := "user=user password=pass host=localhost port=5432 dbname=trains sslmode=disable"
	if connectionString != expected {
		t.Fatalf("expected %q, got %q", expected, connectionString)
	}
}

func TestLoadEnvironment(t *testing.T) {
	t.Setenv("PORT", "8888")
	t.Setenv("ENVIRONMENT", "development")
	t.Setenv("DATABASE_USERNAME", "alice")
	t.Setenv("DATABASE_PASSWORD", "secret")
	t.Setenv("DATABASE_HOST", "db")
	t.Setenv("DATABASE_PORT", "5432")
	t.Setenv("DATABASE_NAME", "api")

	environment := LoadEnvironment()

	if environment.Port != "8888" {
		t.Fatalf("expected port 8888, got %q", environment.Port)
	}

	if !environment.IsDevelopment() {
		t.Fatalf("expected environment to be development")
	}

	if environment.DatabaseConnectionString == "" {
		t.Fatalf("expected database connection string to be populated")
	}

	if environment.EmailAppPassword != "" {
		t.Fatalf("expected empty email app password when secret is unavailable")
	}

	if environment.OpenRouterAPIKey != "" {
		t.Fatalf("expected empty open router api key when secret is unavailable")
	}
}
