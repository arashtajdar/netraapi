package handlers

import (
	"encoding/json"
	"net/http"

	"netra-api/config"
)

func TriviaReward(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := r.Context().Value("user_id").(int)

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

	query := `UPDATE users SET virtual_coins = virtual_coins + ? WHERE id = ?`
	_, err := config.DB.Exec(query, input.CoinsReward, userID)
	if err != nil {
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Coins securely awarded"})
}
