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
			// Express Backend (Core Data & Webhooks)
			{PathPrefix: "/api/v1/problems", TargetURL: getEnv("EXPRESS_BACKEND_URL", "")},
			{PathPrefix: "/api/v1/user", TargetURL: getEnv("EXPRESS_BACKEND_URL", "")},
			{PathPrefix: "/api/webhooks", TargetURL: getEnv("EXPRESS_BACKEND_URL", "")},

			// Go Orchestrator (Workflows)
			{PathPrefix: "/api/v1/evaluations", TargetURL: getEnv("ORCHESTRATOR_URL", "")},
			{PathPrefix: "/api/v1/sessions", TargetURL: getEnv("ORCHESTRATOR_URL", "")},

			// LangGraph Agent (AI & Streaming)
			{PathPrefix: "/api/v1/mentor", TargetURL: getEnv("AGENT_API_URL", "")},
		},
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
