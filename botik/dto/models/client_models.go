package models

import "fmt"

type RawGeocodingDTO struct {
	Name       string            `json:"name"`
	Lat        float64           `json:"lat"`
	Lon        float64           `json:"lon"`
	Country    string            `json:"country"`
	State      string            `json:"state"`
	LocalNames map[string]string `json:"local_names"`
}

type Coord struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

type Weather []struct {
	Main        string `json:"main"`
	Description string `json:"description"`
}

type MainInfo struct {
	Temp      float64 `json:"temp"`
	FeelsLike float64 `json:"feels_like"`
	Pressure  int     `json:"pressure"`
	Humidity  int     `json:"humidity"`
}

type Wind struct {
	Speed float64 `json:"speed"`
}

type Sys struct {
	Country string `json:"country"`
}

type RawWeatherDTO struct {
	Name    string   `json:"name"`
	Coord   Coord    `json:"coord"`
	Weather Weather  `json:"weather"`
	Main    MainInfo `json:"main"`
	Wind    Wind     `json:"wind"`
	Sys     Sys      `json:"sys"`
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
