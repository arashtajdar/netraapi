package handlers

import (
	"encoding/json"
	"net/http"
	"netra-api/config"
	"netra-api/models"
)

func GetSettings(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT setting_key, setting_value FROM app_settings")
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
