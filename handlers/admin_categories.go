package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
)

// validCategoryTables is a whitelist that maps safe type names to their
// corresponding database table names. This prevents SQL injection by ensuring
// user input never reaches SQL string interpolation without validation.
var validCategoryTables = map[string]string{
	"movie":   "movie_categories",
	"series":  "series_categories",
	"live_tv": "live_tv_categories",
	"sports":  "sports_categories",
	"music":   "music_categories",
}

type AdminCategoriesHandler struct {
	db *sql.DB
}

func NewAdminCategoriesHandler(db *sql.DB) *AdminCategoriesHandler {
	return &AdminCategoriesHandler{db: db}
}

func (h *AdminCategoriesHandler) View(w http.ResponseWriter, r *http.Request) {
	categoriesByType := make(map[string][]map[string]interface{})

	for t, tableName := range validCategoryTables {
		rows, err := h.db.Query(fmt.Sprintf("SELECT id, name, slug FROM %s ORDER BY name ASC", tableName))
		if err == nil {
			var cats []map[string]interface{}
			for rows.Next() {
				var id int
				var name, slug string
				rows.Scan(&id, &name, &slug)
				cats = append(cats, map[string]interface{}{
					"ID":   id,
					"Name": name,
					"Slug": slug,
				})
			}
			rows.Close()
			categoriesByType[t] = cats
		}
	}

	renderTemplate(w, "admin_categories.html", map[string]interface{}{
		"CategoriesByType": categoriesByType,
	})
}

func (h *AdminCategoriesHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	t := r.FormValue("type")
	name := r.FormValue("name")

	tableName, ok := validCategoryTables[t]
	if !ok || name == "" {
		http.Error(w, "Invalid category type or missing name", http.StatusBadRequest)
		return
	}

	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	h.db.Exec(fmt.Sprintf("INSERT IGNORE INTO %s (name, slug) VALUES (?, ?)", tableName), name, slug)
	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}

func (h *AdminCategoriesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	t := r.FormValue("type")
	id := r.FormValue("id")

	tableName, ok := validCategoryTables[t]
	if !ok || id == "" {
		http.Error(w, "Invalid category type or missing ID", http.StatusBadRequest)
		return
	}

	h.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName), id)
	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}
