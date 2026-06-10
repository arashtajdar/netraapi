package handlers

import (
	"database/sql"
	"net/http"

	"sheedbox-api/config"

	"github.com/go-chi/chi/v5"
)

type AdminFeaturedHandler struct {
	db *sql.DB
}

func NewAdminFeaturedHandler(db *sql.DB) *AdminFeaturedHandler {
	return &AdminFeaturedHandler{db: db}
}

func (h *AdminFeaturedHandler) View(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT f.id, f.content_type, f.content_id, f.custom_description, f.image_url, f.created_at,
		CASE f.content_type
			WHEN 'movie' THEN (SELECT title FROM movies WHERE id = f.content_id)
			WHEN 'series' THEN (SELECT title FROM series WHERE id = f.content_id)
			WHEN 'live_tv' THEN (SELECT name FROM live_tv_channels WHERE id = f.content_id)
			WHEN 'sports' THEN (SELECT title FROM sports_events WHERE id = f.content_id)
		END as content_title
		FROM featured_items f
		ORDER BY f.created_at DESC
	`
	rows, err := h.db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id, contentId int
		var contentType, created string
		var customDesc, contentTitle, imageUrl sql.NullString

		err := rows.Scan(&id, &contentType, &contentId, &customDesc, &imageUrl, &created, &contentTitle)
		if err == nil {
			items = append(items, map[string]interface{}{
				"ID":                id,
				"ContentType":       contentType,
				"ContentID":         contentId,
				"CustomDescription": customDesc.String,
				"ImageURL":          imageUrl.String,
				"ContentTitle":      contentTitle.String,
				"CreatedAt":         created,
			})
		}
	}
	renderTemplate(w, "admin_featured.html", map[string]interface{}{"Featured": items})
}

func (h *AdminFeaturedHandler) FormView(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin_featured_form.html", nil)
}

func (h *AdminFeaturedHandler) Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	contentType := r.FormValue("content_type")
	contentId := r.FormValue("content_id")
	customDesc := r.FormValue("custom_description")
	imageUrl := r.FormValue("image_url")

	_, err = h.db.Exec("INSERT INTO featured_items (content_type, content_id, custom_description, image_url) VALUES (?, ?, NULLIF(?,''), NULLIF(?,''))", contentType, contentId, customDesc, imageUrl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	config.ClearCachePattern("featured_items")
	http.Redirect(w, r, "/admin/featured", http.StatusSeeOther)
}

func (h *AdminFeaturedHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec("DELETE FROM featured_items WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	config.ClearCachePattern("featured_items")
	http.Redirect(w, r, "/admin/featured", http.StatusSeeOther)
}
