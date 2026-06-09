package middleware

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func init() {
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("supersecretkey_change_in_prod")
	}
}

// JWTMiddleware validates the bearer token and injects user_id into context
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, "Bearer ")
		if len(parts) != 2 {
			http.Error(w, "Invalid Authorization Header Format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or Expired JWT Token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Inject user_id securely into request context
		userID := int(claims["user_id"].(float64))
		ctx := context.WithValue(r.Context(), "user_id", userID)
		
		profileIDStr := r.Header.Get("X-Profile-ID")
		if profileIDStr != "" {
			profileID, err := strconv.Atoi(profileIDStr)
			if err == nil {
				ctx = context.WithValue(ctx, "profile_id", profileID)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
