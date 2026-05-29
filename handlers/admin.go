package handlers

import (
	"html/template"
	"net/http"

	"netra-api/config"
)

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles("views/layout.html", "views/"+tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.ExecuteTemplate(w, "layout.html", data)
}

func AdminDashboardView(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin_dashboard.html", nil)
}

func AdminMoviesView(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin_movies.html", nil)
}

func AdminMoviesFormView(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin_movies_form.html", nil)
}

func AdminSeriesView(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin_series.html", nil)
}

func AdminLiveTVView(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin_livetv.html", nil)
}

func AdminSportsView(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin_sports.html", nil)
}

func AdminMoviesCreate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	releaseDate := r.FormValue("release_date")
	director := r.FormValue("director")
	imdbRating := r.FormValue("imdb_rating")
	localRating := r.FormValue("local_rating")
	posterUrl := r.FormValue("poster_url")
	backdropUrl := r.FormValue("backdrop_url")
	introStart := r.FormValue("intro_start")
	introEnd := r.FormValue("intro_end")
	videoSources := r.FormValue("video_sources")
	subtitles := r.FormValue("subtitles")

	if videoSources == "" {
		videoSources = "[]"
	}
	if subtitles == "" {
		subtitles = "{}"
	}

	query := `INSERT INTO movies (title, description, release_date, director, imdb_rating, local_rating, poster_url, backdrop_url, intro_start, intro_end, video_sources, subtitles) 
			  VALUES (?, ?, NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), ?, ?)`

	_, err = config.DB.Exec(query, title, description, releaseDate, director, imdbRating, localRating, posterUrl, backdropUrl, introStart, introEnd, videoSources, subtitles)
	if err != nil {
		http.Error(w, "Failed to insert movie: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/movies", http.StatusSeeOther)
}

func AdminLiveTVFormView(w http.ResponseWriter, r *http.Request) { renderTemplate(w, "admin_livetv_form.html", nil) }

func AdminLiveTVCreate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.FormValue("name")
	streamUrl := r.FormValue("stream_url")
	logoUrl := r.FormValue("logo_url")

	_, err := config.DB.Exec(`INSERT INTO live_tv_channels (name, stream_url, logo_url) VALUES (?, ?, NULLIF(?,''))`, name, streamUrl, logoUrl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/live-tv", http.StatusSeeOther)
}

func AdminSportsFormView(w http.ResponseWriter, r *http.Request) { renderTemplate(w, "admin_sports_form.html", nil) }

func AdminSportsCreate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	title := r.FormValue("title")
	desc := r.FormValue("description")
	isLive := r.FormValue("is_live") == "true"
	liveUrl := r.FormValue("live_stream_url")
	vidSrc := r.FormValue("video_sources")
	startTime := r.FormValue("start_time")
	
	if vidSrc == "" { vidSrc = "[]" }
	
	_, err := config.DB.Exec(`INSERT INTO sports_events (title, description, is_live, live_stream_url, video_sources, start_time) 
		VALUES (?, NULLIF(?,''), ?, NULLIF(?,''), ?, NULLIF(?,''))`, title, desc, isLive, liveUrl, vidSrc, startTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/sports", http.StatusSeeOther)
}

func AdminSeriesFormView(w http.ResponseWriter, r *http.Request) { renderTemplate(w, "admin_series_form.html", nil) }

func AdminSeriesCreate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	title := r.FormValue("title")
	desc := r.FormValue("description")
	director := r.FormValue("director")
	rating := r.FormValue("rating")
	posterUrl := r.FormValue("poster_url")
	backdropUrl := r.FormValue("backdrop_url")
	cast := r.FormValue("cast_members")
	if cast == "" { cast = "[]" }

	_, err := config.DB.Exec(`INSERT INTO series (title, description, director, rating, poster_url, backdrop_url, cast_members) 
		VALUES (?, NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), ?)`, 
		title, desc, director, rating, posterUrl, backdropUrl, cast)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/series", http.StatusSeeOther)
}
