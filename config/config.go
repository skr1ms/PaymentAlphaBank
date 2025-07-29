package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BaseURL  string
	Username string
	Password string
	IsTest   bool
	Port     string
	DbConfig DbConfig
}

type DbConfig struct {
	URL string
}

// Создание конфигурации для тестовой среды
func NewTestConfig() *Config {
	godotenv.Load()
	return &Config{
		BaseURL:  "https://alfa.rbsuat.com",
		Username: os.Getenv("ALFA_BANK_USERNAME_TEST"),
		Password: os.Getenv("ALFA_BANK_PASSWORD_TEST"),
		IsTest:   true,
		Port:     "3000",	
		DbConfig: DbConfig{
			URL: os.Getenv("DB_URL"),
		},
	}
}

// Создание конфигурации для продакшена
func NewProdConfig() *Config {
	godotenv.Load()
	return &Config{
		BaseURL:  os.Getenv("ALFA_BANK_PROD_URL"),
		Username: os.Getenv("ALFA_BANK_USERNAME"),
		Password: os.Getenv("ALFA_BANK_PASSWORD"),
		IsTest:   false,
		Port:     "3000",
		DbConfig: DbConfig{
			URL: os.Getenv("DB_URL"),
		},
	}
}

