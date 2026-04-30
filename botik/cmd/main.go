package main

import (
	"bot/internal/bot"
	"bot/internal/client"
	"bot/internal/config"
	"bot/internal/services"
)

func main() {
	cfg := config.MustParseConfig()

	httpClient := client.NewWeatherApiClient(cfg.WeatherApiKey)

	service := services.NewWeatherService(httpClient)

	telegramBot := bot.NewBot(cfg.BotApiKey, service)

	telegramBot.StartBot()
}
