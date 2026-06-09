package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
)

type AdminSettingsHandler struct {
	db *sql.DB
}

func NewAdminSettingsHandler(db *sql.DB) *AdminSettingsHandler {
	return &AdminSettingsHandler{db: db}
}

func (h *AdminSettingsHandler) View(w http.ResponseWriter, r *http.Request) {
	var upNextTimer, frontendMenu string
	h.db.QueryRow("SELECT setting_value FROM app_settings WHERE setting_key = 'up_next_timer'").Scan(&upNextTimer)
	h.db.QueryRow("SELECT setting_value FROM app_settings WHERE setting_key = 'frontend_menu'").Scan(&frontendMenu)

	// Remove surrounding quotes for timer if present
	if len(upNextTimer) >= 2 && upNextTimer[0] == '"' && upNextTimer[len(upNextTimer)-1] == '"' {
		upNextTimer = upNextTimer[1 : len(upNextTimer)-1]
	}

	data := map[string]interface{}{
		"UpNextTimer":  upNextTimer,
		"FrontendMenu": frontendMenu,
	}
	renderTemplate(w, "admin_settings.html", data)
}

func (h *AdminSettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	timer := r.FormValue("up_next_timer")
	menu := r.FormValue("frontend_menu")

	h.db.Exec("UPDATE app_settings SET setting_value = ? WHERE setting_key = 'up_next_timer'", fmt.Sprintf(`"%s"`, timer))
	h.db.Exec("UPDATE app_settings SET setting_value = ? WHERE setting_key = 'frontend_menu'", menu)

	http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
}
