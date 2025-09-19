package config

import (
	"fmt"
	"os"
)

// Config holds basic runtime configuration.
type Config struct {
	Port        string
	Env         string
	DB          DBConfig
	Version     string
	JWTSecret   string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// Load returns configuration read from environment variables.
func Load() Config {
	port := os.Getenv("GO_BACKEND_PORT")
	if port == "" { port = "8080" }
	env := os.Getenv("ENV")
	if env == "" { env = "development" }
	db := DBConfig{
		Host:     firstNonEmpty(os.Getenv("POSTGRES_HOST"), "localhost"),
		Port:     firstNonEmpty(os.Getenv("POSTGRES_PORT"), "5432"),
		User:     firstNonEmpty(os.Getenv("POSTGRES_USER"), "codyssey"),
		Password: firstNonEmpty(os.Getenv("POSTGRES_PASSWORD"), "codyssey"),
		Name:     firstNonEmpty(os.Getenv("POSTGRES_DB"), "codyssey"),
		SSLMode:  firstNonEmpty(os.Getenv("POSTGRES_SSLMODE"), "disable"),
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" { jwtSecret = "dev-secret-change-me" }
	return Config{Port: port, Env: env, DB: db, JWTSecret: jwtSecret}
}

func (d DBConfig) ConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values { if v != "" { return v } }
	return ""
}
