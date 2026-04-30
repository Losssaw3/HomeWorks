package services

import (
	"bot/internal/client"
)

type weatherService struct {
	client *client.WeatherApiClient
}

func (w *weatherService) GetWeatherByCoords(lat, long string) (*WeatherResponce, error) {
	weather, err := w.client.GetWeatherByCoords(lat, long)
	if err != nil {
		return nil, err
	}
	return &WeatherResponce{Report: weather.String(), Prefix: "Вот что нашлось по вашему запросу\n"}, nil
}

func (w *weatherService) GetWeatherByCity(cityName, countryCode string) (*WeatherResponce, error) {
	weather, err := w.client.GetCoords(cityName, countryCode)
	if err != nil {
		return nil, err
	}
	return &WeatherResponce{Report: weather.String(), Prefix: "Вот что нашлось по вашему запросу\n"}, nil
}

func NewWeatherService(client *client.WeatherApiClient) *weatherService {
	return &weatherService{client: client}
}

type WeatherResponce struct {
	Prefix string
	Report string
}
