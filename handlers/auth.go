package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"netra-api/config"
	"netra-api/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func init() {
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("supersecretkey_change_in_prod")
	}
}

func Register(w http.ResponseWriter, r *http.Request) {
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

	// SQL Injection prevented using parameterized query (?)
	query := `INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`
	res, err := config.DB.Exec(query, user.Username, user.Email, hash)
	if err != nil {
		http.Error(w, `{"error": "Email or Username already taken"}`, http.StatusConflict)
		return
	}

	id, _ := res.LastInsertId()
	user.ID = int(id)
	user.PasswordHash = ""
	user.VirtualCoins = 500

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "Invalid payload"}`, http.StatusBadRequest)
		return
	}

	var user models.User
	var hash string

	query := `SELECT id, username, email, password_hash, virtual_coins FROM users WHERE email = ?`
	err := config.DB.QueryRow(query, input.Email).Scan(&user.ID, &user.Username, &user.Email, &hash, &user.VirtualCoins)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(input.Password)); err != nil {
		http.Error(w, `{"error": "Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, _ := token.SignedString(jwtSecret)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": tokenString,
		"user":  user,
	})
}

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var input struct {
		IDToken string `json:"idToken"`
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

	payload, err := idtoken.Validate(context.Background(), input.IDToken, clientID)
	if err != nil {
		http.Error(w, `{"error": "Invalid Google token"}`, http.StatusUnauthorized)
		return
	}

	email, ok := payload.Claims["email"].(string)
	if !ok {
		http.Error(w, `{"error": "Email not found in token"}`, http.StatusBadRequest)
		return
	}

	name := email
	if n, ok := payload.Claims["name"].(string); ok {
		name = n
	}

	var user models.User
	var hash string

	query := `SELECT id, username, email, password_hash, virtual_coins FROM users WHERE email = ?`
	err = config.DB.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &hash, &user.VirtualCoins)

	if err == sql.ErrNoRows {
		// Register user
		queryInsert := `INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`
		res, errInsert := config.DB.Exec(queryInsert, name, email, "") // Allow empty hash for google users
		if errInsert != nil {
			http.Error(w, `{"error": "Failed to create user"}`, http.StatusInternalServerError)
			return
		}
		id, _ := res.LastInsertId()
		user.ID = int(id)
		user.Username = name
		user.Email = email
		user.VirtualCoins = 500
	} else if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, _ := token.SignedString(jwtSecret)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": tokenString,
		"user":  user,
	})
}
