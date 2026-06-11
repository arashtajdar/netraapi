package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"sheedbox-api/config"
	"sheedbox-api/models"
	"sheedbox-api/services"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

type AuthHandler struct {
	userService *services.UserService
}

func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{userService: userService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error": "Encryption Failed"}`, http.StatusInternalServerError)
		return
	}

	user.PasswordHash = string(hash)
	err = h.userService.CreateUser(r.Context(), &user)
	if err != nil {
		http.Error(w, `{"error": "Email or Username already taken"}`, http.StatusConflict)
		return
	}

	user.PasswordHash = ""
	user.VirtualCoins = 500

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "Invalid payload"}`, http.StatusBadRequest)
		return
	}

	user, hash, err := h.userService.GetUserByEmail(r.Context(), input.Email)
	if err != nil || user == nil {
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(input.Password)); err != nil {
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    user.ID,
		"user_level": user.UserLevel,
		"exp":        time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, _ := token.SignedString(config.JWTSecret())

	user.PasswordHash = ""
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": tokenString,
		"user":  user,
	})
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var input struct {
		IDToken     string `json:"idToken"`
		AccessToken string `json:"accessToken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "Invalid payload"}`, http.StatusBadRequest)
		return
	}

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		http.Error(w, `{"error": "Google Client ID not configured"}`, http.StatusInternalServerError)
		return
	}

	var email string
	var name string

	if input.IDToken != "" {
		payload, err := idtoken.Validate(context.Background(), input.IDToken, clientID)
		if err != nil {
			http.Error(w, `{"error": "Invalid Google token"}`, http.StatusUnauthorized)
			return
		}

		e, ok := payload.Claims["email"].(string)
		if !ok {
			http.Error(w, `{"error": "Email not found in token"}`, http.StatusBadRequest)
			return
		}
		email = e
		name = email
		if n, ok := payload.Claims["name"].(string); ok {
			name = n
		}
	} else if input.AccessToken != "" {
		req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
		req.Header.Set("Authorization", "Bearer "+input.AccessToken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			http.Error(w, `{"error": "Invalid Google access token"}`, http.StatusUnauthorized)
			return
		}
		defer resp.Body.Close()

		var userInfo struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			http.Error(w, `{"error": "Failed to decode user info"}`, http.StatusInternalServerError)
			return
		}
		email = userInfo.Email
		name = userInfo.Name
		if name == "" {
			name = email
		}
	} else {
		http.Error(w, `{"error": "No token provided"}`, http.StatusBadRequest)
		return
	}

	user, _, err := h.userService.GetUserByEmail(r.Context(), email)
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	if user == nil {
		// Register user
		user = &models.User{
			Username:     name,
			Email:        email,
			PasswordHash: "", // Allow empty hash for Google users
		}
		err = h.userService.CreateUser(r.Context(), user)
		if err != nil {
			http.Error(w, `{"error": "Failed to create user"}`, http.StatusInternalServerError)
			return
		}
		user.VirtualCoins = 500
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    user.ID,
		"user_level": user.UserLevel,
		"exp":        time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, _ := token.SignedString(config.JWTSecret())

	user.PasswordHash = ""
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": tokenString,
		"user":  user,
	})
}
