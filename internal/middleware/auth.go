package middleware

import (
	"bytes"
	"crypto/subtle"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
	svix "github.com/svix/svix-webhooks/go"
)

type AuthMiddleware struct {
	jwks   *keyfunc.JWKS
	issuer string
}

func NewAuthMiddleware(jwksURL, issuer string) (*AuthMiddleware, error) {
	options := keyfunc.Options{}
	jwks, err := keyfunc.Get(jwksURL, options)
	if err != nil {
		return nil, err
	}

	return &AuthMiddleware{
		jwks:   jwks,
		issuer: issuer,
	}, nil
}

func (am *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/webhooks") {

			// 1. Clerk Webhooks (Svix Signature Verification)
			if strings.HasPrefix(r.URL.Path, "/api/webhooks/clerk") {
				secret := os.Getenv("CLERK_WEBHOOK_SECRET")
				if secret == "" {
					slog.Error("CLERK_WEBHOOK_SECRET is not set")
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}

				wh, err := svix.NewWebhook(secret)
				if err != nil {
					slog.Error("Failed to create Svix webhook", "error", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}

				// Limit body size to 1MB to prevent OOM
				r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Payload too large or read error", http.StatusRequestEntityTooLarge)
					return
				}

				// Verify cryptographic signature
				err = wh.Verify(bodyBytes, r.Header)
				if err != nil {
					slog.Warn("Clerk webhook verification failed", "error", err)
					http.Error(w, "Unauthorized: invalid signature", http.StatusUnauthorized)
					return
				}

				// Restore body so proxy can forward it
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				next.ServeHTTP(w, r)
				return
			}

			// 2. LangGraph Webhooks (Query Secret Verification)
			if strings.HasPrefix(r.URL.Path, "/api/webhooks/evaluations") {
				secret := os.Getenv("LANGGRAPH_WEBHOOK_SECRET")
				if secret == "" {
					slog.Error("LANGGRAPH_WEBHOOK_SECRET is not set")
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}

				providedSecret := r.URL.Query().Get("secret")

				// Use ConstantTimeCompare to mitigate timing attacks
				if subtle.ConstantTimeCompare([]byte(providedSecret), []byte(secret)) != 1 {
					slog.Warn("LangGraph webhook verification failed")
					http.Error(w, "Unauthorized: invalid secret", http.StatusUnauthorized)
					return
				}

				next.ServeHTTP(w, r)
				return
			}

			// Strict block for unknown webhooks
			http.Error(w, "Unauthorized: unknown webhook route", http.StatusUnauthorized)
			return
		}

		// --- Standard Clerk JWT Verification ---
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Unauthorized: invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]

		token, err := jwt.Parse(tokenStr, am.jwks.Keyfunc)
		if err != nil {
			slog.Warn("Failed to parse or verify token", "error", err)
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "Unauthorized: token is not valid", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if iss, ok := claims["iss"].(string); !ok || iss != am.issuer {
				http.Error(w, "Unauthorized: invalid issuer", http.StatusUnauthorized)
				return
			}

			// Extract subject (user ID) and inject into request headers for downstream
			if sub, ok := claims["sub"].(string); ok {
				r.Header.Set("X-User-Id", sub)
			}
		} else {
			http.Error(w, "Unauthorized: invalid claims", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
