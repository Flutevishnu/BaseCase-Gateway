package config

import (
	"os"
	"strconv"
)

type Route struct {
	PathPrefix string
	TargetURL  string
}

type Config struct {
	Port         int
	RedisAddr    string
	ClerkJWKSURL string
	ClerkIssuer  string
	Routes       []Route
}

func Load() Config {
	port, _ := strconv.Atoi(getEnv("PORT", "8080"))

	return Config{
		Port:         port,
		RedisAddr:    getEnv("REDIS_ADDR", ""),
		ClerkJWKSURL: getEnv("CLERK_JWKS_URL", ""),
		ClerkIssuer:  getEnv("CLERK_ISSUER", ""),
		Routes: []Route{
			{PathPrefix: "/api/v1/express/", TargetURL: getEnv("EXPRESS_BACKEND_URL", "")},
			{PathPrefix: "/api/v1/orchestrator/", TargetURL: getEnv("ORCHESTRATOR_URL", "")},
			{PathPrefix: "/api/v1/agent/", TargetURL: getEnv("AGENT_API_URL", "")},
		},
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
