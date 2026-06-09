package handlers

import (
	"context"
	"net/http"
	"time"

	"sheedbox-api/config"
)

// LivenessCheck returns a simple 200 OK to indicate the app is running.
func LivenessCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "alive"}`))
}

// ReadinessCheck verifies connections to the database and Redis cache.
func ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// Check Database
	if err := config.DB.PingContext(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status": "readyz_failed", "database": "disconnected"}`))
		return
	}

	// Check Redis (if active)
	if config.RedisClient != nil {
		if err := config.RedisClient.Ping(ctx).Err(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status": "readyz_failed", "redis": "disconnected"}`))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ready", "database": "connected", "redis": "connected"}`))
}
