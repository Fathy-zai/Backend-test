package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type weatherstack struct {
	cli *http.Client
	key string
}

func NewWeatherstack(cli *http.Client, key string) Provider {
	return &weatherstack{cli: cli, key: key}
}

func (w *weatherstack) Name() string { return "weatherstack" }

func (w *weatherstack) Get(ctx context.Context, city string) (Measurement, error) {
	// API: http://api.weatherstack.com/current?access_key=KEY&query=Melbourne
	u := fmt.Sprintf("http://api.weatherstack.com/current?access_key=%s&query=%s",
		url.QueryEscape(w.key), url.QueryEscape(city))

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	resp, err := w.cli.Do(req)
	if err != nil {
		return Measurement{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Measurement{}, fmt.Errorf("status %d", resp.StatusCode)
	}

	var payload struct {
		Success *bool                  `json:"success,omitempty"`
		Error   *struct{ Info string } `json:"error,omitempty"`
		Current struct {
			Temperature float64 `json:"temperature"`
			WindSpeed   float64 `json:"wind_speed"` // km/h per docs
		} `json:"current"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Measurement{}, err
	}
	if payload.Error != nil {
		return Measurement{}, errors.New(payload.Error.Info)
	}

	return Measurement{TempC: payload.Current.Temperature, WindKph: payload.Current.WindSpeed}, nil
}
