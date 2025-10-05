package main

import (
	"log"
	"net/http"
	"os"

	"weather-service/internal/config"
	"weather-service/internal/httpserver"
	"weather-service/internal/providers"
)

func main() {
	cfg := config.FromEnv()

	httpClient := &http.Client{Timeout: cfg.HTTPTimeout}

	// TOM: Minor: This could be pulled into a LoadProviders() fn and that would make it easier to test this part of the initialisation
	var provs []providers.Provider
	if k := os.Getenv("WEATHERSTACK_KEY"); k != "" {
		provs = append(provs, providers.NewWeatherstack(httpClient, k))
	}
	if k := os.Getenv("OPENWEATHER_KEY"); k != "" {
		provs = append(provs, providers.NewOpenWeather(httpClient, k))
	}

	// TOM: It would be nice to know which providers couldn't be loaded rather than a general message. That would make debugging easier
	if len(provs) == 0 {
		log.Fatal("no providers configured; set WEATHERSTACK_KEY and/or OPENWEATHER_KEY")
	}

	mux := httpserver.NewMux(cfg, provs)

	log.Printf("listening on :%d", cfg.Port)
	if err := http.ListenAndServe(cfg.Addr(), mux); err != nil {
		log.Fatal(err)
	}
}
