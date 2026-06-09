package handlers

import (
	"net/http"
	"time"

	"sheedbox-api/config"
	"sheedbox-api/models"
	"sheedbox-api/services"
)

type AdminSportsHandler struct {
	sportsService *services.SportsService
}

func NewAdminSportsHandler(sportsService *services.SportsService) *AdminSportsHandler {
	return &AdminSportsHandler{sportsService: sportsService}
}

func (h *AdminSportsHandler) View(w http.ResponseWriter, r *http.Request) {
	events, err := h.sportsService.ListEvents(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var eventsList []map[string]interface{}
	for _, e := range events {
		var stStr string
		if e.StartTime != nil {
			stStr = e.StartTime.Format("2006-01-02 15:04")
		} else {
			stStr = "N/A"
		}
		eventsList = append(eventsList, map[string]interface{}{
			"ID":        e.ID,
			"Title":     e.Title,
			"IsLive":    e.IsLive,
			"StartTime": stStr,
		})
	}

	renderTemplate(w, "admin_sports.html", map[string]interface{}{"Events": eventsList})
}

func (h *AdminSportsHandler) FormView(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin_sports_form.html", nil)
}

func (h *AdminSportsHandler) Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	desc := r.FormValue("description")
	isLive := r.FormValue("is_live") == "true"
	liveUrl := r.FormValue("live_stream_url")
	vidSrc := r.FormValue("video_sources")
	startTime := r.FormValue("start_time")

	var startTimePtr *time.Time
	if startTime != "" {
		if t, err := time.Parse("2006-01-02T15:04", startTime); err == nil {
			startTimePtr = &t
		} else if t, err := time.Parse("2006-01-02 15:04:05", startTime); err == nil {
			startTimePtr = &t
		}
	}

	videoSources := []byte("[]")
	if vidSrc != "" {
		videoSources = []byte(vidSrc)
	}

	event := models.SportsEvent{
		Title:         title,
		Description:   stringPtr(desc),
		IsLive:        isLive,
		LiveStreamURL: stringPtr(liveUrl),
		VideoSources:  videoSources,
		StartTime:     startTimePtr,
	}

	err = h.sportsService.CreateEvent(r.Context(), &event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	config.ClearCachePattern("sports_*")

	http.Redirect(w, r, "/admin/sports", http.StatusSeeOther)
}
