package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
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
