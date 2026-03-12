package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	APP_ENV string `envconfig:"APP_ENV" default:"dev"`

	DBHost string `envconfig:"DB_HOST"`
	DBPort string `envconfig:"DB_PORT"`
	DBUser string `envconfig:"DB_USER"`
	DBPass string `envconfig:"DB_PASS"`
	DBName string `envconfig:"DB_NAME"`

	JWTSecret                string `envconfig:"JWT_SECRET" default:"your-secret-key-change-this-in-production"`
	JWTExpiresInAccessToken  int    `envconfig:"JWT_EXPIRES_IN_ACCESS_TOKEN" default:"15"`     // minutes
	JWTExpiresInRefreshToken int    `envconfig:"JWT_EXPIRES_IN_REFRESH_TOKEN" default:"10080"` // minutes (7 days)
	RefreshSecret            string `envconfig:"REFRESH_SECRET" default:"your-refresh-secret-key-change-this-in-production"`

	RedisHost     string `envconfig:"REDIS_HOST" default:"localhost"`
	RedisPort     int    `envconfig:"REDIS_PORT" default:"6379"`
	RedisPassword string `envconfig:"REDIS_PASSWORD"`

	OpenAIKey string `envconfig:"OPENAI_KEY"`
}

func LoadConfig() *Config {
	_ = godotenv.Load(".env")

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	_ = godotenv.Load(".env." + env)

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		fmt.Println("ERROR loading config:", err)
		return nil
	}

	return &cfg
}
