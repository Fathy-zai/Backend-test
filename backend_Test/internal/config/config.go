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
	// TOM: Sometimes it makes sense to have defaults, but in the case of our app here it might be better to let it fail when an env var is missing. Reason being is if we accidentally mess up our config we will find out very quickly, i.e. while we're working on the system rather than in prod. Take a look at the Virtual Accounts API to see how we handled that
	port := getenvInt("PORT", 8080)
	ttl := getenvInt("CACHE_TTL_MS", 3000)
	to := getenvInt("HTTP_TIMEOUT_MS", 1500)
	return Config{
		Port:        port,
		CacheTTL:    time.Duration(ttl) * time.Millisecond,
		HTTPTimeout: time.Duration(to) * time.Millisecond,
	}
}

// TOM: Nitpick: go follows camel casing, i.e. getEnvInt()
func getenvInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
