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
	// TOM: Go's built-in multiplexer is sufficient for simple APIs like this one but it becomes a bit unwieldy the more routes you add. Consider using the chi library for future extensions as this makes the routing really easy
	mux := http.NewServeMux()
	// TOM: As is, this is pretty difficult to unit test because all the handlers are defined here rather than in main. In the Virtual Accounts API, we create an object to hold the handlers and pass that into the server setup fn
	mux.HandleFunc("/v1/weather", s.handleWeather)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})
	return logging(mux)
}

func (s *server) handleWeather(w http.ResponseWriter, r *http.Request) {
	// TOM: Makes sense to set the default like this while getting the WIP going, but in the final product I'd expect to see defaults like this exposed for configuration at a higher level, e.g. in main or more likely, the config file
	city := r.URL.Query().Get("city")
	if city == "" {
		city = "melbourne"
	}
	city = strings.ToLower(city)

	// Serve fresh cache if available
	// TOM: Nice work on the cache
	if v, ok := s.cache.GetFresh(city); ok {
		writeJSON(w, v, http.StatusOK)
		return
	}

	// Query providers with deadline
	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.HTTPTimeout)
	defer cancel()

	// TOM: Works for few providers and consumers but probably wouldn't scale well. Consider the case in which you've got 10 different providers, each of which is slow and unreliable and you're trying to serve 10 requests per second. Those requests are likely to get processed really slowly if the first few providers in the list are slow. The implementation here might even contribute to the slowness because you'll be sending all your traffic to only one (the first) provider. Consider setting up a simple load balancer that distributes the requests evenly and keeps track of unhealthy servers rather than just looping through the providers.
	for _, p := range s.providers {
		m, err := p.Get(ctx, city)
		if err != nil {
			log.Printf("provider %s error: %v", p.Name(), err)
			// try next provider with a fresh timeout
			ctx, cancel = context.WithTimeout(r.Context(), s.cfg.HTTPTimeout)
			defer cancel()
			continue
		}
		// TOM: Nice one on setting up a model. Minor: this model serves the responsibility of being an internal data structure as well as modelling the outgoing request. Consider splitting them into separate models. See Virtual Accounts for inspiration. There, we have separate models for incoming and outgoing requests, as well one for the database
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

// TOM: Would be good to leave a comment to explain why this is needed
func round1(f float64) float64 {
	return float64(int(f*10+0.5)) / 10
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Consider using Go's slog library for logging with built in time stamps and formatting
		start := time.Now()
		next.ServeHTTP(w, r)
		// TOM: Minor: I'd put this before we serve the request in case it fails. That ensures we've always got something to look at when debugging
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
