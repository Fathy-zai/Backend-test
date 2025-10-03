package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// OpenWeather metric units: temp in °C, wind.speed in m/s. Convert wind to km/h.
const mpsToKph = 3.6

type openweather struct {
	cli *http.Client
	key string
}

func NewOpenWeather(cli *http.Client, key string) Provider {
	return &openweather{cli: cli, key: key}
}

func (o *openweather) Name() string { return "openweather" }

func (o *openweather) Get(ctx context.Context, city string) (Measurement, error) {
	// API: http://api.openweathermap.org/data/2.5/weather?q=melbourne,AU&appid=KEY&units=metric
	q := url.Values{}
	q.Set("q", city+",AU")
	q.Set("appid", o.key)
	q.Set("units", "metric")

	u := url.URL{Scheme: "https", Host: "api.openweathermap.org", Path: "/data/2.5/weather", RawQuery: q.Encode()}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	resp, err := o.cli.Do(req)
	if err != nil {
		return Measurement{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// read a small body for debugging
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return Measurement{}, fmt.Errorf("status %d body: %s", resp.StatusCode, string(b))
	}

	var payload struct {
		Main struct {
			Temp float64 `json:"temp"`
		} `json:"main"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Measurement{}, err
	}

	return Measurement{TempC: payload.Main.Temp, WindKph: payload.Wind.Speed * mpsToKph}, nil
}
