package handlers

import (
	"encoding/json"
	"net/http"
	"sheedbox-api/config"
	"sheedbox-api/models"
	"github.com/go-chi/chi/v5"
	"strconv"
)

func GetUserProfiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := r.Context().Value("user_id").(int)

	query := `SELECT id, user_id, name, avatar_url, is_kids_mode, created_at FROM user_profiles WHERE user_id = ?`
	rows, err := config.DB.Query(query, userID)
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var profiles []models.UserProfile
	for rows.Next() {
		var p models.UserProfile
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.AvatarURL, &p.IsKidsMode, &p.CreatedAt); err == nil {
			profiles = append(profiles, p)
		}
	}
	if profiles == nil {
		profiles = []models.UserProfile{}
	}
	json.NewEncoder(w).Encode(profiles)
}

func CreateProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := r.Context().Value("user_id").(int)

	var req models.UserProfile
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error": "Name is required"}`, http.StatusBadRequest)
		return
	}

	query := `INSERT INTO user_profiles (user_id, name, avatar_url, is_kids_mode) VALUES (?, ?, ?, ?)`
	res, err := config.DB.Exec(query, userID, req.Name, req.AvatarURL, req.IsKidsMode)
	if err != nil {
		http.Error(w, `{"error": "Failed to create profile"}`, http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	req.ID = int(id)
	req.UserID = userID
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := r.Context().Value("user_id").(int)
	profileID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error": "Invalid profile ID"}`, http.StatusBadRequest)
		return
	}

	var req models.UserProfile
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	query := `UPDATE user_profiles SET name = ?, avatar_url = ?, is_kids_mode = ? WHERE id = ? AND user_id = ?`
	_, err = config.DB.Exec(query, req.Name, req.AvatarURL, req.IsKidsMode, profileID, userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to update profile"}`, http.StatusInternalServerError)
		return
	}

	req.ID = profileID
	req.UserID = userID
	json.NewEncoder(w).Encode(req)
}

func DeleteProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := r.Context().Value("user_id").(int)
	profileID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error": "Invalid profile ID"}`, http.StatusBadRequest)
		return
	}

	query := `DELETE FROM user_profiles WHERE id = ? AND user_id = ?`
	_, err = config.DB.Exec(query, profileID, userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to delete profile"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
