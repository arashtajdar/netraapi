package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	adminCookieName = "sheedbox_admin_session"
	adminCookieMaxAge = 12 * 60 * 60 // 12 hours in seconds
)

var (
	adminUsername string
	adminPassword string
	adminSessionSecret []byte
	adminLoginTmpl *template.Template
)

// InitAdminAuth loads admin credentials from the environment.
// If ADMIN_USERNAME or ADMIN_PASSWORD are not set, the admin panel will be inaccessible.
func InitAdminAuth() {
	adminUsername = os.Getenv("ADMIN_USERNAME")
	adminPassword = os.Getenv("ADMIN_PASSWORD")

	if adminUsername == "" || adminPassword == "" {
		log.Println("WARNING: ADMIN_USERNAME and/or ADMIN_PASSWORD not set. Admin panel will be inaccessible.")
	}

	// Use JWT_SECRET as the base for admin session signing, with a domain separator
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "fallback-admin-secret-not-for-production"
	}
	adminSessionSecret = []byte("admin-session:" + secret)

	// Parse login template at startup
	var err error
	adminLoginTmpl, err = template.ParseFiles("views/admin_login.html")
	if err != nil {
		log.Printf("WARNING: Could not parse admin login template: %v", err)
	}
}

// AdminAuthMiddleware protects admin routes with cookie-based session authentication.
// It allows /admin/login through (both GET and POST) and redirects everything else
// to the login page if no valid session cookie is present.
func AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow the login route through without authentication
		if strings.HasSuffix(r.URL.Path, "/login") || strings.HasSuffix(r.URL.Path, "/logout") {
			next.ServeHTTP(w, r)
			return
		}

		// Check for valid admin session cookie
		cookie, err := r.Cookie(adminCookieName)
		if err != nil || !validateAdminSession(cookie.Value) {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AdminLoginView renders the admin login page.
func AdminLoginView(w http.ResponseWriter, r *http.Request) {
	if adminLoginTmpl == nil {
		http.Error(w, "Login template not found", http.StatusInternalServerError)
		return
	}
	adminLoginTmpl.Execute(w, map[string]interface{}{
		"Error": "",
	})
}

// AdminLoginSubmit handles the login form submission.
func AdminLoginSubmit(w http.ResponseWriter, r *http.Request) {
	if adminUsername == "" || adminPassword == "" {
		http.Error(w, "Admin credentials not configured on the server.", http.StatusServiceUnavailable)
		return
	}

	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Constant-time comparison to prevent timing attacks
	usernameMatch := hmac.Equal([]byte(username), []byte(adminUsername))
	passwordMatch := hmac.Equal([]byte(password), []byte(adminPassword))

	if !usernameMatch || !passwordMatch {
		if adminLoginTmpl != nil {
			adminLoginTmpl.Execute(w, map[string]interface{}{
				"Error": "Invalid username or password.",
			})
			return
		}
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create signed session token
	sessionToken := createAdminSession()

	http.SetCookie(w, &http.Cookie{
		Name:     adminCookieName,
		Value:    sessionToken,
		Path:     "/admin",
		MaxAge:   adminCookieMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
	})

	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

// AdminLogout clears the admin session cookie.
func AdminLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     adminCookieName,
		Value:    "",
		Path:     "/admin",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

// createAdminSession creates a signed session token containing an expiration timestamp.
// Format: "expiresUnix.hmacHex"
func createAdminSession() string {
	expires := time.Now().Add(time.Duration(adminCookieMaxAge) * time.Second).Unix()
	payload := fmt.Sprintf("%d", expires)
	signature := signPayload(payload)
	return fmt.Sprintf("%s.%s", payload, signature)
}

// validateAdminSession checks that the session token is valid and not expired.
func validateAdminSession(token string) bool {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return false
	}

	payload := parts[0]
	signature := parts[1]

	// Verify HMAC signature
	expected := signPayload(payload)
	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return false
	}

	// Check expiration
	var expires int64
	if _, err := fmt.Sscanf(payload, "%d", &expires); err != nil {
		return false
	}

	return time.Now().Unix() < expires
}

// signPayload computes an HMAC-SHA256 of the payload using the admin session secret.
func signPayload(payload string) string {
	mac := hmac.New(sha256.New, adminSessionSecret)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// AdminSessionJSON returns the admin session status as JSON (for AJAX checks).
func AdminSessionJSON(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(adminCookieName)
	valid := err == nil && validateAdminSession(cookie.Value)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"authenticated": valid})
}
