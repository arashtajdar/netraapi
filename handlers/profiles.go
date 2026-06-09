package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"sheedbox-api/contextkeys"
	"sheedbox-api/models"
	"sheedbox-api/services"

	"github.com/go-chi/chi/v5"
)

type ProfileHandler struct {
	profileService *services.ProfileService
}

func NewProfileHandler(profileService *services.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

func (h *ProfileHandler) GetUserProfiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := contextkeys.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	profiles, err := h.profileService.GetProfiles(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error": "Database retrieval failed"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(profiles)
}

func (h *ProfileHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := contextkeys.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req models.UserProfile
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error": "Name is required"}`, http.StatusBadRequest)
		return
	}

	req.UserID = userID
	err := h.profileService.CreateProfile(r.Context(), &req)
	if err != nil {
		http.Error(w, `{"error": "Failed to create profile"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := contextkeys.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
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

	req.ID = profileID
	req.UserID = userID
	err = h.profileService.UpdateProfile(r.Context(), &req)
	if err != nil {
		http.Error(w, `{"error": "Failed to update profile"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(req)
}

func (h *ProfileHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := contextkeys.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	profileID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error": "Invalid profile ID"}`, http.StatusBadRequest)
		return
	}

	err = h.profileService.DeleteProfile(r.Context(), profileID, userID)
	if err != nil {
		http.Error(w, `{"error": "Failed to delete profile"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
