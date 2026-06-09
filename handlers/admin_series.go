package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"sheedbox-api/config"
	"sheedbox-api/models"
	"sheedbox-api/services"
)

type AdminSeriesHandler struct {
	seriesService *services.SeriesService
}

func NewAdminSeriesHandler(seriesService *services.SeriesService) *AdminSeriesHandler {
	return &AdminSeriesHandler{seriesService: seriesService}
}

func (h *AdminSeriesHandler) View(w http.ResponseWriter, r *http.Request) {
	seriesList, err := h.seriesService.ListSeries(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var seriesData []map[string]interface{}
	for _, s := range seriesList {
		var rating float64
		if s.Rating != nil {
			rating = *s.Rating
		}
		seriesData = append(seriesData, map[string]interface{}{
			"ID":     s.ID,
			"Title":  s.Title,
			"Rating": rating,
		})
	}

	renderTemplate(w, "admin_series.html", map[string]interface{}{"Series": seriesData})
}

func (h *AdminSeriesHandler) FormView(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin_series_form.html", nil)
}

func (h *AdminSeriesHandler) Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	desc := r.FormValue("description")
	director := r.FormValue("director")
	rating := r.FormValue("rating")
	posterUrl := r.FormValue("poster_url")
	backdropUrl := r.FormValue("backdrop_url")
	cast := r.FormValue("cast_members")

	var ratingPtr *float64
	if rating != "" {
		if val, err := strconv.ParseFloat(rating, 64); err == nil {
			ratingPtr = &val
		}
	}

	castJSON := []byte("[]")
	if cast != "" {
		castJSON = []byte(cast)
	}

	series := models.Series{
		Title:       title,
		Description: desc,
		Director:    stringPtr(director),
		Rating:      ratingPtr,
		PosterURL:   stringPtr(posterUrl),
		BackdropURL: stringPtr(backdropUrl),
		CastMembers: json.RawMessage(castJSON),
	}

	err = h.seriesService.CreateSeries(r.Context(), &series)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	config.ClearCachePattern("series_*")

	http.Redirect(w, r, "/admin/series", http.StatusSeeOther)
}
