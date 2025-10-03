package httpserver

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"weather-service/internal/cache"
	"weather-service/internal/config"
	"weather-service/internal/model"
	"weather-service/internal/providers"
)

type server struct {
	cfg   config.Config
	cache interface {
		GetFresh(string) (model.Weather, bool)
		GetStale(string) (model.Weather, bool)
		Set(string, model.Weather)
	}
	providers []providers.Provider
}

func NewMux(cfg config.Config, provs []providers.Provider) http.Handler {
	s := &server{cfg: cfg, cache: cache.NewMemory(cfg.CacheTTL), providers: provs}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/weather", s.handleWeather)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})
	return logging(mux)
}

func (s *server) handleWeather(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		city = "melbourne"
	}
	city = strings.ToLower(city)

	// Serve fresh cache if available
	if v, ok := s.cache.GetFresh(city); ok {
		writeJSON(w, v, http.StatusOK)
		return
	}

	// Query providers with deadline
	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.HTTPTimeout)
	defer cancel()

	for _, p := range s.providers {
		m, err := p.Get(ctx, city)
		if err != nil {
			log.Printf("provider %s error: %v", p.Name(), err)
			// try next provider with a fresh timeout
			ctx, cancel = context.WithTimeout(r.Context(), s.cfg.HTTPTimeout)
			defer cancel()
			continue
		}
		out := model.Weather{WindSpeedKph: round1(m.WindKph), TemperatureCelsius: round1(m.TempC)}
		s.cache.Set(city, out)
		writeJSON(w, out, http.StatusOK)
		return
	}

	// All providers failed: try stale cache
	if v, ok := s.cache.GetStale(city); ok {
		writeJSON(w, v, http.StatusOK)
		return
	}

	http.Error(w, "all providers unavailable", http.StatusBadGateway)
}

func writeJSON(w http.ResponseWriter, v any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func round1(f float64) float64 {
	return float64(int(f*10+0.5)) / 10
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
