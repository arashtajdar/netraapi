package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"sheedbox-api/models"
)

type SettingsHandler struct {
	db *sql.DB
}

func NewSettingsHandler(db *sql.DB) *SettingsHandler {
	return &SettingsHandler{db: db}
}

func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(), "SELECT setting_key, setting_value FROM app_settings")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	settings := make(map[string]json.RawMessage)
	for rows.Next() {
		var s models.AppSetting
		if err := rows.Scan(&s.SettingKey, &s.SettingValue); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		settings[s.SettingKey] = s.SettingValue
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}
