package client

import (
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

type RawGeocodingDTO struct {
	Name       string            `json:"name"`
	Lat        float64           `json:"lat"`
	Lon        float64           `json:"lon"`
	Country    string            `json:"country"`
	State      string            `json:"state"`
	LocalNames map[string]string `json:"local_names"`
}

type RawWeatherDTO struct {
	Name  string `json:"name"`
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	Sys struct {
		Country string `json:"country"`
	} `json:"sys"`
}

func (r *RawWeatherDTO) String() string {
	description := "нет данных"
	if len(r.Weather) > 0 {
		description = r.Weather[0].Description
	}

	return fmt.Sprintf(
		"Погода в %s (%s):\n"+
			"- Состояние: %s\n"+
			"- Температура: %.1f°C (ощущается как %.1f°C)\n"+
			"- Влажность: %d%%\n"+
			"- Ветер: %.1f м/с",
		r.Name, r.Sys.Country,
		description,
		r.Main.Temp, r.Main.FeelsLike,
		r.Main.Humidity,
		r.Wind.Speed,
	)
}

func NewWeatherApiClient(apiKey string) *WeatherApiClient {
	return &WeatherApiClient{client: &http.Client{Timeout: 5 * time.Second}, apiKey: apiKey}
}

func (w *WeatherApiClient) GetWeatherByCoords(lat, lon string) (*RawWeatherDTO, error) {
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

	var result *RawWeatherDTO
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (w *WeatherApiClient) GetCoords(city, countryCode string) (*RawWeatherDTO, error) {
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

	var results []RawGeocodingDTO
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
	// Возвращаем первый найденный результат
	return weather, nil
}
