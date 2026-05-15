package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	AppPort         string
	RedisAddr       string
	JWTSecret       string
	DBHost          string
	DBPort          string
	DBUsername      string
	DBPassword      string
	DBName          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func MustLoad() *AppConfig {
	err := godotenv.Load()

	if err != nil {
		log.Println("No .env file found")
	}

	accessTokenTTL, _ := time.ParseDuration(os.Getenv("ACCESS_TOKEN_TTL"))
	refreshTokenTTL, _ := time.ParseDuration(os.Getenv("REFRESH_TOKEN_TTL"))
	return &AppConfig{
		AppPort:         os.Getenv("APP_PORT"),
		RedisAddr:       os.Getenv("REDIS_ADDR"),
		DBHost:          os.Getenv("DB_HOST"),
		DBPort:          os.Getenv("DB_PORT"),
		DBUsername:      os.Getenv("DB_USER_NAME"),
		DBPassword:      os.Getenv("DB_PASSWORD"),
		DBName:          os.Getenv("DB_NAME"),
		AccessTokenTTL:  accessTokenTTL,
		RefreshTokenTTL: refreshTokenTTL,
	}
}
