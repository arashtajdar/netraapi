package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

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

	res, err := config.DB.Exec(`INSERT INTO live_tv_channels (name, stream_url, logo_url, youtube_url) VALUES (?, NULLIF(?,''), NULLIF(?,''), NULLIF(?,''))`, name, streamUrl, logoUrl, youtubeUrl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	channelID, _ := res.LastInsertId()
	saveEPG(channelID, r.FormValue("epg_data"))

	http.Redirect(w, r, "/admin/live-tv", http.StatusSeeOther)
}

type EPGData struct {
	ProgramTitle string `json:"program_title"`
	Description  string `json:"description"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
}

func saveEPG(channelID int64, epgJSON string) {
	if epgJSON == "" || epgJSON == "[]" {
		config.DB.Exec("DELETE FROM epg WHERE channel_id = ?", channelID)
		return
	}
	var epg []EPGData
	err := json.Unmarshal([]byte(epgJSON), &epg)
	if err != nil {
		log.Println("Invalid EPG JSON:", err)
		return
	}
	config.DB.Exec("DELETE FROM epg WHERE channel_id = ?", channelID)
	for _, item := range epg {
		// Convert ISO8601 strings to MySQL expected format or leave to driver
		var startStr, endStr interface{}
		if t, err := time.Parse(time.RFC3339, item.StartTime); err == nil {
			startStr = t.Format("2006-01-02 15:04:05")
		} else {
			startStr = item.StartTime // fallback
		}
		if t, err := time.Parse(time.RFC3339, item.EndTime); err == nil {
			endStr = t.Format("2006-01-02 15:04:05")
		} else {
			endStr = item.EndTime // fallback
		}

		config.DB.Exec("INSERT INTO epg (channel_id, program_title, description, start_time, end_time) VALUES (?, ?, ?, ?, ?)",
			channelID, item.ProgramTitle, item.Description, startStr, endStr)
	}
}

func getEPG(channelID int) string {
	rows, err := config.DB.Query("SELECT program_title, description, start_time, end_time FROM epg WHERE channel_id = ? ORDER BY start_time ASC", channelID)
	if err != nil {
		return "[]"
	}
	defer rows.Close()
	var epg []EPGData
	for rows.Next() {
		var item EPGData
		var start, end []byte
		var desc sql.NullString
		if err := rows.Scan(&item.ProgramTitle, &desc, &start, &end); err == nil {
			item.Description = desc.String
			// start and end are []byte from MySQL timestamp
			tStart, _ := time.Parse("2006-01-02 15:04:05", string(start))
			tEnd, _ := time.Parse("2006-01-02 15:04:05", string(end))
			item.StartTime = tStart.Format(time.RFC3339)
			item.EndTime = tEnd.Format(time.RFC3339)
			if item.StartTime == "0001-01-01T00:00:00Z" {
				item.StartTime = string(start)
			}
			if item.EndTime == "0001-01-01T00:00:00Z" {
				item.EndTime = string(end)
			}
			epg = append(epg, item)
		}
	}
	if len(epg) == 0 {
		return "[]"
	}
	b, _ := json.MarshalIndent(epg, "", "  ")
	return string(b)
}

func AdminLiveTVEditFormView(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var channel map[string]interface{}
	var name, streamUrl, logoUrl, youtubeUrl sql.NullString
	err := config.DB.QueryRow("SELECT id, name, stream_url, logo_url, youtube_url FROM live_tv_channels WHERE id = ?", id).Scan(&id, &name, &streamUrl, &logoUrl, &youtubeUrl)
	if err == nil {
		channelID := 0
		fmt.Sscanf(id, "%d", &channelID)
		epgData := getEPG(channelID)
		channel = map[string]interface{}{
			"ID":         id,
			"Name":       name.String,
			"StreamURL":  streamUrl.String,
			"LogoURL":    logoUrl.String,
			"YoutubeURL": youtubeUrl.String,
			"EPGData":    epgData,
		}
	} else {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}
	renderTemplate(w, "admin_livetv_form.html", map[string]interface{}{"Channel": channel})
}

func AdminLiveTVUpdate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	r.ParseForm()
	name := r.FormValue("name")
	streamUrl := r.FormValue("stream_url")
	logoUrl := r.FormValue("logo_url")
	youtubeUrl := r.FormValue("youtube_url")

	_, err := config.DB.Exec(`UPDATE live_tv_channels SET name=?, stream_url=NULLIF(?,''), logo_url=NULLIF(?,''), youtube_url=NULLIF(?,'') WHERE id=?`, name, streamUrl, logoUrl, youtubeUrl, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	var channelID int64
	fmt.Sscanf(id, "%d", &channelID)
	saveEPG(channelID, r.FormValue("epg_data"))

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
