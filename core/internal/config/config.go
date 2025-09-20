package config

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type HTTPServer struct {
	Host string
	Port string
}

type RedisCache struct {
	Host     string
	Port     string
	Password string
}

type Postgres struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type TelegramBot struct {
	Token  string
	Domain string
}

type Config struct {
	HTTP        HTTPServer
	Redis       RedisCache
	Postgres    Postgres
	TelegramBot TelegramBot
}

const logtag = "[config]"

func Load() *Config {
	configPath := flag.String("config", "", "path env file")
	flag.Parse()

	if *configPath != "" {
		if err := godotenv.Load(*configPath); err != nil {
			log.Fatalf("%s err loading env from file : %v", logtag, err)
		}
		log.Printf("%s using env from : %s", logtag, *configPath)
	} else {
		log.Printf("%s using env from .env", logtag)
		_ = godotenv.Load()
	}

	cfg := &Config{
		HTTP:        *newHTTP(),
		Redis:       *newRedis(),
		Postgres:    *newPostgres(),
		TelegramBot: *newTelegramBot(),
	}

	log.Printf("%s backend config : %+v\n", logtag, cfg)
	return cfg
}

func newHTTP() *HTTPServer {
	return &HTTPServer{
		Port: getenv("HTTP_PORT", "8080"),
		Host: getenv("HTTP_HOST", "localhost"),
	}
}

func newRedis() *RedisCache {
	return &RedisCache{
		Port:     getenv("REDIS_PORT", "6379"),
		Host:     getenv("REDIS_HOST", "redis"),
		Password: getenv("REDIS_PASSWORD", "shared"),
	}
}

func newPostgres() *Postgres {
	return &Postgres{
		Host:     getenv("DB_HOST", "localhost"),
		Port:     getenv("DB_PORT", "5432"),
		User:     getenv("DB_USER", "admin"),
		Password: getenv("DB_PASSWORD", "shared"),
		DBName:   getenv("DB_NAME", "test"),
		SSLMode:  getenv("DB_SSLMODE", "disable"),
	}
}

func newTelegramBot() *TelegramBot {
	return &TelegramBot{
		Token:  getenv("TELEGRAM_TOKEN", ""),
		Domain: getenv("TELEGRAM_DOMAIN", ""),
	}
}

func getenv(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		fmt.Printf("%s %s undefined. Using default value %s\n", logtag, key, defaultValue)
		return defaultValue
	}
	fmt.Printf("%s %s = %s\n", logtag, key, val)
	return val
}
