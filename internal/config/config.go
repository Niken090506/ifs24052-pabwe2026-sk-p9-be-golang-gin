package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv, AppPort, AppKey                  string
	AdminUsername, AdminPassword             string
	DBConnection, DBHost, DBPort, DBDatabase string
	DBUsername, DBPassword, DBSSLMode        string
	SearchEndpoint                           string
	OpenAIEndpoint, OpenAIKey, OpenAIModel   string
}

func get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// LoadConfig membaca .env (jika ada) lalu environment dengan nilai default.
func LoadConfig() *Config {
	_ = godotenv.Load()
	return &Config{
		AppEnv:         get("APP_ENV", "development"),
		AppPort:        get("APP_PORT", "8081"),
		AppKey:         get("APP_KEY", "delcom_blog_secret_key_super_secure_2026_random_string_32chars"),
		AdminUsername:  get("ADMIN_USERNAME", "admin"),
		AdminPassword:  get("ADMIN_PASSWORD", "admin"),
		DBConnection:   get("DB_CONNECTION", "pgsql"),
		DBHost:         get("DB_HOST", "127.0.0.1"),
		DBPort:         get("DB_PORT", "5432"),
		DBDatabase:     get("DB_DATABASE", "pabwe_p9_blog"),
		DBUsername:     get("DB_USERNAME", "postgres"),
		DBPassword:     get("DB_PASSWORD", "postgres"),
		DBSSLMode:      get("DB_SSLMODE", "disable"),
		SearchEndpoint: get("SEARCH_ENGINE_ENDPOINT", "https://delcom.org/api/v1/search"),
		OpenAIEndpoint: get("OPENAI_ENDPOINT", "https://delcom.org/api/v1"),
		OpenAIKey:      get("OPENAI_KEY", "key"),
		OpenAIModel:    get("OPENAI_MODEL", "nama-model"),
	}
}
