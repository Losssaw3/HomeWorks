package config

import (
	"os"

	"github.com/joho/godotenv"
)

const (
	cfgPatg = "./config/.env"
)

type Config struct {
	BotApiKey     string
	WeatherApiKey string
}

func MustParseConfig() *Config {
	err := godotenv.Load(cfgPatg)
	if err != nil {
		panic("Error loading .env file")
	}
	return &Config{BotApiKey: os.Getenv("TG_BOT_TOKEN"), WeatherApiKey: os.Getenv("WAEATHER_API")}
}
