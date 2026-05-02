package services

import (
	"bot/dto/models"
	"bot/internal/client"
)

type WeatherServiceI interface {
	GetWeatherByCoords(lat, long string) (*models.WeatherResponce, error)
	GetWeatherByCity(cityName, countryCode string) (*models.WeatherResponce, error)
}

type weatherService struct {
	client *client.WeatherApiClient
}

func NewWeatherService(client *client.WeatherApiClient) *weatherService {
	return &weatherService{client: client}
}

func (w *weatherService) GetWeatherByCoords(lat, long string) (*models.WeatherResponce, error) {
	weather, err := w.client.GetWeatherByCoords(lat, long)
	if err != nil {
		return nil, err
	}
	return &models.WeatherResponce{Report: weather.String(), Prefix: "Вот что нашлось по вашему запросу\n"}, nil
}

func (w *weatherService) GetWeatherByCity(cityName, countryCode string) (*models.WeatherResponce, error) {
	weather, err := w.client.GetCoords(cityName, countryCode)
	if err != nil {
		return nil, err
	}
	return &models.WeatherResponce{Report: weather.String(), Prefix: "Вот что нашлось по вашему запросу\n"}, nil
}
