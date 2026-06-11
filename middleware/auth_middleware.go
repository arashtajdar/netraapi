package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"sheedbox-api/config"
	"sheedbox-api/contextkeys"

	"github.com/golang-jwt/jwt/v5"
)

// JWTMiddleware validates the bearer token and injects user_id into context
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Missing Authorization Header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, "Bearer ")
		if len(parts) != 2 {
			http.Error(w, `{"error": "Invalid Authorization Header Format"}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Use centralized key function that enforces HS256 algorithm
		token, err := jwt.Parse(tokenString, config.JWTKeyFunc)

		if err != nil || !token.Valid {
			http.Error(w, `{"error": "Invalid or Expired JWT Token"}`, http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, `{"error": "Invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, `{"error": "Invalid token: missing user_id"}`, http.StatusUnauthorized)
			return
		}

		// Inject user_id securely into request context using typed keys
		userID := int(userIDFloat)
		ctx := contextkeys.WithUserID(r.Context(), userID)
		w.Header().Set("X-Log-User-ID", strconv.Itoa(userID))

		userLevelFloat, ok := claims["user_level"].(float64)
		userLevel := 1
		if ok {
			userLevel = int(userLevelFloat)
		}
		ctx = contextkeys.WithUserLevel(ctx, userLevel)

		profileIDStr := r.Header.Get("X-Profile-ID")
		if profileIDStr != "" {
			profileID, err := strconv.Atoi(profileIDStr)
			if err == nil {
				ctx = contextkeys.WithProfileID(ctx, profileID)
				w.Header().Set("X-Log-Profile-ID", profileIDStr)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalJWTMiddleware parses the bearer token if present and injects user context,
// but does NOT block the request if the token is missing or invalid.
func OptionalJWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(authHeader, "Bearer ")
		if len(parts) != 2 {
			next.ServeHTTP(w, r)
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, config.JWTKeyFunc)

		if err != nil || !token.Valid {
			next.ServeHTTP(w, r)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		userID := int(userIDFloat)
		ctx := contextkeys.WithUserID(r.Context(), userID)

		userLevelFloat, ok := claims["user_level"].(float64)
		userLevel := 1
		if ok {
			userLevel = int(userLevelFloat)
		}
		ctx = contextkeys.WithUserLevel(ctx, userLevel)

		profileIDStr := r.Header.Get("X-Profile-ID")
		if profileIDStr != "" {
			profileID, err := strconv.Atoi(profileIDStr)
			if err == nil {
				ctx = contextkeys.WithProfileID(ctx, profileID)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
