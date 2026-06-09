package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"sheedbox-api/config"
	"sheedbox-api/models"
	"sheedbox-api/services"
)

type AdminMovieHandler struct {
	movieService *services.MovieService
}

func NewAdminMovieHandler(movieService *services.MovieService) *AdminMovieHandler {
	return &AdminMovieHandler{movieService: movieService}
}

func (h *AdminMovieHandler) View(w http.ResponseWriter, r *http.Request) {
	movies, err := h.movieService.ListMovies(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var movieData []map[string]interface{}
	for _, m := range movies {
		var director string
		if m.Director != nil {
			director = *m.Director
		}
		var rating float64
		if m.IMDBRating != nil {
			rating = *m.IMDBRating
		}
		movieData = append(movieData, map[string]interface{}{
			"ID":       m.ID,
			"Title":    m.Title,
			"Director": director,
			"Rating":   rating,
		})
	}

	renderTemplate(w, "admin_movies.html", map[string]interface{}{"Movies": movieData})
}

func (h *AdminMovieHandler) FormView(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin_movies_form.html", nil)
}

func (h *AdminMovieHandler) Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	releaseDate := r.FormValue("release_date")
	director := r.FormValue("director")
	imdbRatingStr := r.FormValue("imdb_rating")
	localRatingStr := r.FormValue("local_rating")
	posterUrl := r.FormValue("poster_url")
	backdropUrl := r.FormValue("backdrop_url")
	introStartStr := r.FormValue("intro_start")
	introEndStr := r.FormValue("intro_end")
	videoUrl := r.FormValue("video_url")
	subtitles := r.FormValue("subtitles")

	var imdbRating *float64
	if imdbRatingStr != "" {
		if val, err := strconv.ParseFloat(imdbRatingStr, 64); err == nil {
			imdbRating = &val
		}
	}

	var localRating *float64
	if localRatingStr != "" {
		if val, err := strconv.ParseFloat(localRatingStr, 64); err == nil {
			localRating = &val
		}
	}

	var introStart *int
	if introStartStr != "" {
		if val, err := strconv.Atoi(introStartStr); err == nil {
			introStart = &val
		}
	}

	var introEnd *int
	if introEndStr != "" {
		if val, err := strconv.Atoi(introEndStr); err == nil {
			introEnd = &val
		}
	}

	videoSources := []byte("[]")
	if videoUrl != "" {
		videoSources = []byte(fmt.Sprintf(`[{"quality": "Original", "url": "%s"}]`, videoUrl))
	}

	subtitlesJSON := []byte("{}")
	if subtitles != "" {
		subtitlesJSON = []byte(subtitles)
	}

	movie := models.Movie{
		Title:        title,
		Description:  description,
		ReleaseDate:  stringPtr(releaseDate),
		Director:     stringPtr(director),
		IMDBRating:   imdbRating,
		LocalRating:  localRating,
		PosterURL:    stringPtr(posterUrl),
		BackdropURL:  stringPtr(backdropUrl),
		IntroStart:   introStart,
		IntroEnd:     introEnd,
		VideoSources: videoSources,
		Subtitles:    subtitlesJSON,
	}

	err = h.movieService.CreateMovie(r.Context(), &movie)
	if err != nil {
		http.Error(w, "Failed to insert movie: "+err.Error(), http.StatusInternalServerError)
		return
	}

	config.ClearCachePattern("movies_*")

	http.Redirect(w, r, "/admin/movies", http.StatusSeeOther)
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
