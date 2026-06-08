package handlers

import (
	"database/sql"
	"net/http"
	"sheedbox-api/config"

	"github.com/go-chi/chi/v5"
)

func AdminFeaturedView(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT f.id, f.content_type, f.content_id, f.custom_description, f.created_at,
		CASE f.content_type
			WHEN 'movie' THEN (SELECT title FROM movies WHERE id = f.content_id)
			WHEN 'series' THEN (SELECT title FROM series WHERE id = f.content_id)
			WHEN 'live_tv' THEN (SELECT name FROM live_tv_channels WHERE id = f.content_id)
			WHEN 'sports' THEN (SELECT title FROM sports_events WHERE id = f.content_id)
		END as content_title
		FROM featured_items f
		ORDER BY f.created_at DESC
	`
	rows, err := config.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id, contentId int
		var contentType, created string
		var customDesc, contentTitle sql.NullString

		err := rows.Scan(&id, &contentType, &contentId, &customDesc, &created, &contentTitle)
		if err == nil {
			items = append(items, map[string]interface{}{
				"ID":                id,
				"ContentType":       contentType,
				"ContentID":         contentId,
				"CustomDescription": customDesc.String,
				"ContentTitle":      contentTitle.String,
				"CreatedAt":         created,
			})
		}
	}
	renderTemplate(w, "admin_featured.html", map[string]interface{}{"Featured": items})
}

func AdminFeaturedFormView(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin_featured_form.html", nil)
}

func AdminFeaturedCreate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	contentType := r.FormValue("content_type")
	contentId := r.FormValue("content_id")
	customDesc := r.FormValue("custom_description")

	_, err := config.DB.Exec("INSERT INTO featured_items (content_type, content_id, custom_description) VALUES (?, ?, NULLIF(?,''))", contentType, contentId, customDesc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/featured", http.StatusSeeOther)
}

func AdminFeaturedDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := config.DB.Exec("DELETE FROM featured_items WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/featured", http.StatusSeeOther)
}
