package model

type Weather struct {
	WindSpeedKph       float64 `json:"wind_speed"`
	TemperatureCelsius float64 `json:"temperature_degrees"`
}
