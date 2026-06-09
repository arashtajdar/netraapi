package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"sheedbox-api/contextkeys"
)

// responseWriterWrapper intercepts the status code and bytes written to log them.
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode    int
	bytesWritten  int
	headerWritten bool
}

func newResponseWriterWrapper(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	if rw.headerWritten {
		return
	}
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
	rw.headerWritten = true
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	if !rw.headerWritten {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// generateRequestID generates a secure random string for request tracking.
func generateRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "unknown-req-id"
	}
	return hex.EncodeToString(b)
}

// RequestLogger is a middleware that assigns a Request ID and logs structured HTTP metrics.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = generateRequestID()
		}

		w.Header().Set("X-Request-ID", reqID)

		// Inject Request ID into Context
		ctx := contextkeys.WithRequestID(r.Context(), reqID)
		r = r.WithContext(ctx)

		wrapper := newResponseWriterWrapper(w)
		startTime := time.Now()

		defer func() {
			duration := time.Since(startTime)
			
			userIDStr := wrapper.Header().Get("X-Log-User-ID")
			profileIDStr := wrapper.Header().Get("X-Log-Profile-ID")

			wrapper.Header().Del("X-Log-User-ID")
			wrapper.Header().Del("X-Log-Profile-ID")

			// Base attributes
			attrs := []slog.Attr{
				slog.String("request_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_ip", r.RemoteAddr),
				slog.Int("status", wrapper.statusCode),
				slog.Duration("duration", duration),
				slog.Int("bytes", wrapper.bytesWritten),
			}

			if userIDStr != "" {
				if uid, err := strconv.Atoi(userIDStr); err == nil {
					attrs = append(attrs, slog.Int("user_id", uid))
				}
			}
			if profileIDStr != "" {
				if pid, err := strconv.Atoi(profileIDStr); err == nil {
					attrs = append(attrs, slog.Int("profile_id", pid))
				}
			}

			// Choose log level based on response code
			level := slog.LevelInfo
			if wrapper.statusCode >= 500 {
				level = slog.LevelError
			} else if wrapper.statusCode >= 400 {
				level = slog.LevelWarn
			}

			slog.LogAttrs(
				r.Context(),
				level,
				"HTTP Request",
				attrs...,
			)
		}()

		next.ServeHTTP(wrapper, r)
	})
}
