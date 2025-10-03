package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port        int
	CacheTTL    time.Duration
	HTTPTimeout time.Duration
}

func (c Config) Addr() string { return ":" + strconv.Itoa(c.Port) }

func FromEnv() Config {
	port := getenvInt("PORT", 8080)
	ttl := getenvInt("CACHE_TTL_MS", 3000)
	to := getenvInt("HTTP_TIMEOUT_MS", 1500)
	return Config{
		Port:        port,
		CacheTTL:    time.Duration(ttl) * time.Millisecond,
		HTTPTimeout: time.Duration(to) * time.Millisecond,
	}
}

func getenvInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
