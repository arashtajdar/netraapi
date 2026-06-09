package config

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecretBytes []byte

// InitJWT loads the JWT secret from the environment.
// It panics if JWT_SECRET is not set — this is intentional.
// A missing secret in production would allow token forgery.
func InitJWT() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("FATAL: JWT_SECRET environment variable is required but not set. " +
			"Refusing to start with an insecure default. " +
			"Set JWT_SECRET in your .env or Railway environment variables.")
	}
	if len(secret) < 32 {
		log.Println("WARNING: JWT_SECRET is shorter than 32 characters. Consider using a longer secret for production.")
	}
	jwtSecretBytes = []byte(secret)
}

// JWTSecret returns the JWT signing key. Must be called after InitJWT().
func JWTSecret() []byte {
	if jwtSecretBytes == nil {
		panic("config.JWTSecret() called before config.InitJWT()")
	}
	return jwtSecretBytes
}

// JWTKeyFunc is a reusable jwt.Keyfunc that enforces HS256 signing method.
// Use this in all jwt.Parse calls to prevent algorithm confusion attacks.
func JWTKeyFunc(token *jwt.Token) (interface{}, error) {
	// Enforce HMAC signing method — prevents "alg: none" and RSA key confusion
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return jwtSecretBytes, nil
}
