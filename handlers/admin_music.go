package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"sheedbox-api/models"
	"sheedbox-api/services"

	"github.com/go-chi/chi/v5"
)

type AdminMusicHandler struct {
	musicService *services.MusicService
}

func NewAdminMusicHandler(musicService *services.MusicService) *AdminMusicHandler {
	return &AdminMusicHandler{musicService: musicService}
}

func (h *AdminMusicHandler) View(w http.ResponseWriter, r *http.Request) {
	music, err := h.musicService.ListMusic(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var musicData []map[string]interface{}
	for _, m := range music {
		var artist string
		if m.Artist != nil {
			artist = *m.Artist
		}
		musicData = append(musicData, map[string]interface{}{
			"ID":     m.ID,
			"Title":  m.Title,
			"Artist": artist,
		})
	}

	renderTemplate(w, "admin_music.html", map[string]interface{}{"Music": musicData})
}

func (h *AdminMusicHandler) FormView(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin_music_form.html", nil)
}

func (h *AdminMusicHandler) Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	artist := r.FormValue("artist")
	desc := r.FormValue("description")
	releaseDate := r.FormValue("release_date")
	posterUrl := r.FormValue("poster_url")
	backdropUrl := r.FormValue("backdrop_url")
	videoUrl := r.FormValue("video_url")

	vidSrc := []byte("[]")
	if videoUrl != "" {
		vidSrc = []byte(fmt.Sprintf(`[{"quality": "Original", "url": "%s"}]`, videoUrl))
	}

	music := models.MusicContent{
		Title:        title,
		Description:  stringPtr(desc),
		Artist:       stringPtr(artist),
		ReleaseDate:  stringPtr(releaseDate),
		PosterURL:    stringPtr(posterUrl),
		BackdropURL:  stringPtr(backdropUrl),
		VideoSources: vidSrc,
		AudioSources: []byte("[]"),
	}

	err = h.musicService.CreateMusic(r.Context(), &music)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/music", http.StatusSeeOther)
}

func (h *AdminMusicHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "Missing music ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid music ID", http.StatusBadRequest)
		return
	}

	err = h.musicService.DeleteMusic(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/music", http.StatusSeeOther)
}
