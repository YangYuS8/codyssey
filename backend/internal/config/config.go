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
	AutoMigrate bool
	LogLevel    string
	MaxSubmissionCodeBytes int // 代码长度上限
	MaxRequestBodyBytes    int // 全局请求体限制
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
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" { logLevel = "info" }
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
	autoMig := os.Getenv("AUTO_MIGRATE") == "true"
	maxCode := 128 * 1024 // 128KB 默认
	if v := os.Getenv("MAX_SUBMISSION_CODE_BYTES"); v != "" { if n, err := atoiSafe(v); err == nil && n > 0 { maxCode = n } }
	maxBody := 512 * 1024 // 512KB 默认
	if v := os.Getenv("MAX_REQUEST_BODY_BYTES"); v != "" { if n, err := atoiSafe(v); err == nil && n > 0 { maxBody = n } }
	return Config{Port: port, Env: env, DB: db, JWTSecret: jwtSecret, AutoMigrate: autoMig, LogLevel: logLevel, MaxSubmissionCodeBytes: maxCode, MaxRequestBodyBytes: maxBody}
}

// Validate performs basic sanity checks; panic early if critical settings missing in non-dev.
func (c Config) Validate() error {
    if c.Env != "development" && c.JWTSecret == "dev-secret-change-me" {
        return fmt.Errorf("JWT_SECRET must be set in %s env", c.Env)
    }
    return nil
}

func (d DBConfig) ConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values { if v != "" { return v } }
	return ""
}

func atoiSafe(s string) (int, error) {
	var n int
	for _, ch := range s {
		if ch < '0' || ch > '9' { return 0, fmt.Errorf("invalid int") }
		n = n*10 + int(ch-'0')
	}
	return n, nil
}
