package config

import "os"

// Config holds basic runtime configuration.
type Config struct {
	Port string
}

// Load returns configuration read from environment variables.
func Load() Config {
	port := os.Getenv("GO_BACKEND_PORT")
	if port == "" {
		port = "8080"
	}
	return Config{Port: port}
}
