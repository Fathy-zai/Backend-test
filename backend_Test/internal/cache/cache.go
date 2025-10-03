package cache

import (
	"sync"
	"time"

	"weather-service/internal/model"
)

type entry struct {
	val       model.Weather
	createdAt time.Time
}

type Memory struct {
	mu   sync.RWMutex
	data map[string]entry
	ttl  time.Duration
}

func NewMemory(ttl time.Duration) *Memory {
	return &Memory{data: make(map[string]entry), ttl: ttl}
}

func (m *Memory) GetFresh(key string) (model.Weather, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.data[key]
	if !ok {
		return model.Weather{}, false
	}
	if time.Since(e.createdAt) <= m.ttl {
		return e.val, true
	}
	return model.Weather{}, false
}

// GetStale returns any cached value regardless of their age.
func (m *Memory) GetStale(key string) (model.Weather, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.data[key]
	if !ok {
		return model.Weather{}, false
	}
	return e.val, true
}

func (m *Memory) Set(key string, v model.Weather) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = entry{val: v, createdAt: time.Now()}
}
