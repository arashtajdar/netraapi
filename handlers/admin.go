package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"netra-api/config"

	"github.com/go-chi/chi/v5"
)

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles("views/layout.html", "views/"+tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.ExecuteTemplate(w, "layout.html", data)
}

type DashboardStats struct {
	MoviesCount int
	SeriesCount int
	LiveTVCount int
	SportsCount int
}

func AdminDashboardView(w http.ResponseWriter, r *http.Request) {
	var stats DashboardStats

	config.DB.QueryRow("SELECT COUNT(*) FROM movies").Scan(&stats.MoviesCount)
	config.DB.QueryRow("SELECT COUNT(*) FROM series").Scan(&stats.SeriesCount)
	config.DB.QueryRow("SELECT COUNT(*) FROM live_tv_channels").Scan(&stats.LiveTVCount)
	config.DB.QueryRow("SELECT COUNT(*) FROM sports_events").Scan(&stats.SportsCount)

	renderTemplate(w, "admin_dashboard.html", stats)
}

func AdminMoviesView(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT id, title, director, imdb_rating FROM movies ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var movies []map[string]interface{}
	for rows.Next() {
		var id int
		var title string
		var director sql.NullString
		var rating sql.NullFloat64
		err := rows.Scan(&id, &title, &director, &rating)
		if err == nil {
			movies = append(movies, map[string]interface{}{
				"ID":       id,
				"Title":    title,
				"Director": director.String,
				"Rating":   rating.Float64,
			})
		} else {
			log.Println("Scan error:", err)
		}
	}

	renderTemplate(w, "admin_movies.html", map[string]interface{}{"Movies": movies})
}

func AdminMoviesFormView(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin_movies_form.html", nil)
}

func AdminSeriesView(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT id, title, rating FROM series ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var seriesList []map[string]interface{}
	for rows.Next() {
		var id int
		var title string
		var rating sql.NullFloat64
		err := rows.Scan(&id, &title, &rating)
		if err == nil {
			seriesList = append(seriesList, map[string]interface{}{
				"ID":     id,
				"Title":  title,
				"Rating": rating.Float64,
			})
		}
	}

	renderTemplate(w, "admin_series.html", map[string]interface{}{"Series": seriesList})
}

func AdminLiveTVView(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT id, name, stream_url, logo_url, youtube_url FROM live_tv_channels ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var channels []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		var streamUrl, logoUrl, youtubeUrl sql.NullString
		err := rows.Scan(&id, &name, &streamUrl, &logoUrl, &youtubeUrl)
		if err == nil {
			channels = append(channels, map[string]interface{}{
				"ID":         id,
				"Name":       name,
				"StreamURL":  streamUrl.String,
				"LogoURL":    logoUrl.String,
				"YoutubeURL": youtubeUrl.String,
			})
		}
	}

	renderTemplate(w, "admin_livetv.html", map[string]interface{}{"Channels": channels})
}

func AdminSportsView(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT id, title, is_live, start_time FROM sports_events ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var eventsList []map[string]interface{}
	for rows.Next() {
		var id int
		var title string
		var isLive bool
		var startTime sql.NullTime
		err := rows.Scan(&id, &title, &isLive, &startTime)
		if err == nil {
			var stStr string
			if startTime.Valid {
				stStr = startTime.Time.Format("2006-01-02 15:04")
			} else {
				stStr = "N/A"
			}
			eventsList = append(eventsList, map[string]interface{}{
				"ID":        id,
				"Title":     title,
				"IsLive":    isLive,
				"StartTime": stStr,
			})
		}
	}

	renderTemplate(w, "admin_sports.html", map[string]interface{}{"Events": eventsList})
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
	videoUrl := r.FormValue("video_url")
	subtitles := r.FormValue("subtitles")

	videoSources := "[]"
	if videoUrl != "" {
		videoSources = fmt.Sprintf(`[{"quality": "Original", "url": "%s"}]`, videoUrl)
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
	youtubeUrl := r.FormValue("youtube_url")

	_, err := config.DB.Exec(`INSERT INTO live_tv_channels (name, stream_url, logo_url, youtube_url) VALUES (?, NULLIF(?,''), NULLIF(?,''), NULLIF(?,''))`, name, streamUrl, logoUrl, youtubeUrl)
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

func AdminLiveTVDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing channel ID", http.StatusBadRequest)
		return
	}
	_, err := config.DB.Exec("DELETE FROM live_tv_channels WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Failed to delete channel: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/live-tv", http.StatusSeeOther)
}
