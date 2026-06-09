package handlers

import (
	"encoding/json"
	"net/http"

	"sheedbox-api/contextkeys"
	"sheedbox-api/services"
)

type GamificationHandler struct {
	userService *services.UserService
}

func NewGamificationHandler(userService *services.UserService) *GamificationHandler {
	return &GamificationHandler{userService: userService}
}

func (h *GamificationHandler) TriviaReward(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, ok := contextkeys.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var input struct {
		CoinsReward int `json:"coins_reward"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	// Security validation: Cap the incoming coins string to prevent fraud.
	// In production, we would validate hashes matching specific quiz answers.
	if input.CoinsReward > 50 {
		http.Error(w, `{"error": "Fraud detected"}`, http.StatusBadRequest)
		return
	}

	err := h.userService.AwardCoins(r.Context(), userID, input.CoinsReward)
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Coins securely awarded"})
}
