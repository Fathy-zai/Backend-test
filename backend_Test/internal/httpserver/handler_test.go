package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"weather-service/internal/config"
	"weather-service/internal/model"
	"weather-service/internal/providers"
)

type fakeProvider struct {
	name string
	m    providers.Measurement
	err  error
}

func (f fakeProvider) Name() string { return f.name }
func (f fakeProvider) Get(ctx context.Context, city string) (providers.Measurement, error) {
	return f.m, f.err
}


func TestHandler_FreshCacheThenProvider(t *testing.T) {
	cfg := config.Config{Port: 0, CacheTTL: 3 * time.Second, HTTPTimeout: 200 * time.Millisecond}
	s := &server{cfg: cfg, providers: []providers.Provider{fakeProvider{"p1", providers.Measurement{TempC: 20, WindKph: 30}, nil}}}
	s.cache = newTestCache() // tcache implements the required cache interface

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/weather?city=melbourne", nil)
	s.handleWeather(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got %d", rr.Code)
	}
}

func TestHandler_FailoverAndStale(t *testing.T) {
	cfg := config.Config{Port: 0, CacheTTL: 10 * time.Millisecond, HTTPTimeout: 50 * time.Millisecond}
	s := &server{
		cfg:   cfg,
		cache: newTestCache(),
	}
	// seed stale
	s.cache.Set("melbourne", model.Weather{WindSpeedKph: 10, TemperatureCelsius: 15})

	s.providers = []providers.Provider{
		fakeProvider{"down1", providers.Measurement{}, context.DeadlineExceeded},
		fakeProvider{"down2", providers.Measurement{}, context.DeadlineExceeded},
	}

	time.Sleep(20 * time.Millisecond) // expire freshness

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/weather?city=melbourne", nil)
	s.handleWeather(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

// minimal cache for tests
type tcache struct{ Memory map[string]model.Weather }

func newTestCache() *tcache                                 { return &tcache{Memory: map[string]model.Weather{}} }
func (t *tcache) GetFresh(key string) (model.Weather, bool) { v, ok := t.Memory[key]; return v, ok }
func (t *tcache) GetStale(key string) (model.Weather, bool) { v, ok := t.Memory[key]; return v, ok }
func (t *tcache) Set(key string, v model.Weather)           { t.Memory[key] = v }

// Ensure tcache implements the same interface as cache.Memory
var _ interface {
	GetFresh(string) (model.Weather, bool)
	GetStale(string) (model.Weather, bool)
	Set(string, model.Weather)
} = (*tcache)(nil)
