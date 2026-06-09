package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type AdminUsersHandler struct {
	db *sql.DB
}

func NewAdminUsersHandler(db *sql.DB) *AdminUsersHandler {
	return &AdminUsersHandler{db: db}
}

func (h *AdminUsersHandler) View(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, username, email, virtual_coins, created_at FROM users ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var username, email string
		var coins int
		var createdAt string
		err := rows.Scan(&id, &username, &email, &coins, &createdAt)
		if err == nil {
			users = append(users, map[string]interface{}{
				"ID":           id,
				"Username":     username,
				"Email":        email,
				"VirtualCoins": coins,
				"CreatedAt":    createdAt,
			})
		}
	}

	renderTemplate(w, "admin_users.html", map[string]interface{}{"Users": users})
}

func (h *AdminUsersHandler) EditFormView(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var user map[string]interface{}
	var username, email string
	var coins int
	err := h.db.QueryRow("SELECT id, username, email, virtual_coins FROM users WHERE id = ?", id).Scan(&id, &username, &email, &coins)
	if err == nil {
		user = map[string]interface{}{
			"ID":           id,
			"Username":     username,
			"Email":        email,
			"VirtualCoins": coins,
		}
	} else {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	renderTemplate(w, "admin_users_form.html", map[string]interface{}{"User": user})
}

func (h *AdminUsersHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	coins := r.FormValue("virtual_coins")

	_, err = h.db.Exec("UPDATE users SET username=?, email=?, virtual_coins=? WHERE id=?", username, email, coins, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (h *AdminUsersHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}
	_, err := h.db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Failed to delete user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}
