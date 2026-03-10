package configuration

import (
	"fmt"
	"os"
	"vulpz/train-api/src/secret"
)

type Environment struct {
	Environment              ApplicationEnvironment
	Port                     string
	DatabaseConnectionString string
	EmailAppPassword         string
}

type ApplicationEnvironment string

const (
	EnvironmentDevelopment ApplicationEnvironment = "development"
	EnvironmentProduction  ApplicationEnvironment = "production"
)

func LoadEnvironment() Environment {
	return Environment{
		Port:                     loadPort(),
		Environment:              loadApplicationEnvironment(),
		DatabaseConnectionString: loadDatabaseConnectionString(),
		EmailAppPassword:         loadEmailAppPassword(),
	}
}

func loadPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "8000"
	}
	return port
}

func loadEmailAppPassword() string {
	appPassword, loadError := secret.LoadSecret("email")
	if loadError != nil {
		return ""
	}
	return string(appPassword)
}

func loadApplicationEnvironment() ApplicationEnvironment {
	switch os.Getenv("ENVIRONMENT") {
	case "development":
		return EnvironmentDevelopment
	default:
		return EnvironmentProduction
	}
}

func loadDatabaseConnectionString() string {
	return fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		os.Getenv("DATABASE_USERNAME"),
		os.Getenv("DATABASE_PASSWORD"),
		os.Getenv("DATABASE_HOST"),
		os.Getenv("DATABASE_PORT"),
		os.Getenv("DATABASE_NAME"),
	)
}

func (environment *Environment) IsDevelopment() bool {
	return environment.Environment == EnvironmentDevelopment
}

func (environment *Environment) IsProduction() bool {
	return environment.Environment == EnvironmentProduction
}
