package services

import (
	"bot/dto/models"
)

type WeatherApiClientI interface {
	GetCoords(city, countryCode string) (*models.RawWeatherDTO, error)
	GetWeatherByCoords(lat, lon string) (*models.RawWeatherDTO, error)
}
type weatherService struct {
	client WeatherApiClientI
}

func NewWeatherService(client WeatherApiClientI) *weatherService {
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
