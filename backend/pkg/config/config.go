package config

import (
	"log"
	"os"
)

type Config struct {
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
	DatabaseURL string
	RedisAddr   string
	RedisURL    string
	JWTSecret   string
	Port        string
	CORSOrigins string
}

func Load() *Config {
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" || jwtSecret == "default-secret-change-me" {
		log.Fatal("FATAL: JWT_SECRET must be set to a secure value (not the default)")
	}

	return &Config{
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "quizuser"),
		DBPassword:  getEnv("DB_PASSWORD", "quizpass"),
		DBName:      getEnv("DB_NAME", "quizdb"),
		DBSSLMode:   getEnv("DB_SSLMODE", "disable"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		RedisURL:    getEnv("REDIS_URL", ""),
		JWTSecret:   jwtSecret,
		Port:        getEnv("PORT", "8080"),
		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:5175,http://localhost:3000"),
	}
}

func (c *Config) DSN() string {
	return "host=" + c.DBHost +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" port=" + c.DBPort +
		" sslmode=" + c.DBSSLMode + " TimeZone=UTC"
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
