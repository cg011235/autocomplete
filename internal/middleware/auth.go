package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"golang.org/x/time/rate"
)

var (
	secretKey []byte
	// Create a rate limiter with a rate of 1 request per second and a burst size of 3.
	limiter = rate.NewLimiter(1, 3)
)

// SetSecretKey sets the secret key for JWT authentication.
func SetSecretKey(key []byte) {
	secretKey = key
}

// Define a custom type for context keys to avoid potential conflicts.
type contextKey string

const userContextKey contextKey = "user"

// JwtMiddleware handles JWT authentication.
func JwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		// Remove the "Bearer " prefix from the token string.
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		// Parse and validate the token.
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return secretKey, nil
		})

		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Store the token claims in the context.
		ctx := context.WithValue(r.Context(), userContextKey, token.Claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RateLimitMiddleware handles rate limiting.
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is allowed by the rate limiter.
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
