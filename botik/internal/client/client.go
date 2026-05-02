package client

import (
	"bot/dto/models"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type WeatherApiClient struct {
	client *http.Client
	apiKey string
}

func NewWeatherApiClient(apiKey string) *WeatherApiClient {
	return &WeatherApiClient{client: &http.Client{Timeout: 5 * time.Second}, apiKey: apiKey}
}

func (w *WeatherApiClient) GetWeatherByCoords(lat, lon string) (*models.RawWeatherDTO, error) {
	baseURL := "https://api.openweathermap.org/data/2.5/weather"

	params := url.Values{}
	params.Add("lat", lat)
	params.Add("lon", lon)
	params.Add("units", "metric")
	params.Add("lang", "ru")
	params.Add("appid", w.apiKey)

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := w.client.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result *models.RawWeatherDTO
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (w *WeatherApiClient) GetCoords(city, countryCode string) (*models.RawWeatherDTO, error) {
	baseURL := "http://api.openweathermap.org/geo/1.0/direct"

	params := url.Values{}
	params.Add("q", fmt.Sprintf("%s,%s", city, countryCode))
	params.Add("limit", "1")
	params.Add("appid", w.apiKey)

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := w.client.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocoding api error: status %d", resp.StatusCode)
	}

	var results []models.RawGeocodingDTO
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("city not found: %s", city)
	}
	lat, long := results[0].Lat, results[0].Lon
	latStr := fmt.Sprintf("%v", lat)
	lonStr := fmt.Sprintf("%v", long)
	weather, err := w.GetWeatherByCoords(latStr, lonStr)
	if err != nil {
		return nil, err
	}
	return weather, nil
}
