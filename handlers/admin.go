package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"sheedbox-api/config"

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
	rows, err := config.DB.Query("SELECT id, name, stream_url, logo_url, youtube_url, youtube_channel_url FROM live_tv_channels ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var channels []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		var streamUrl, logoUrl, youtubeUrl, youtubeChannelUrl sql.NullString
		err := rows.Scan(&id, &name, &streamUrl, &logoUrl, &youtubeUrl, &youtubeChannelUrl)
		if err == nil {
			channels = append(channels, map[string]interface{}{
				"ID":                id,
				"Name":              name,
				"StreamURL":         streamUrl.String,
				"LogoURL":           logoUrl.String,
				"YoutubeURL":        youtubeUrl.String,
				"YoutubeChannelURL": youtubeChannelUrl.String,
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
	youtubeChannelUrl := r.FormValue("youtube_channel_url")

	res, err := config.DB.Exec(`INSERT INTO live_tv_channels (name, stream_url, logo_url, youtube_url, youtube_channel_url) VALUES (?, NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''))`, name, streamUrl, logoUrl, youtubeUrl, youtubeChannelUrl)
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
	var name, streamUrl, logoUrl, youtubeUrl, youtubeChannelUrl sql.NullString
	err := config.DB.QueryRow("SELECT id, name, stream_url, logo_url, youtube_url, youtube_channel_url FROM live_tv_channels WHERE id = ?", id).Scan(&id, &name, &streamUrl, &logoUrl, &youtubeUrl, &youtubeChannelUrl)
	if err == nil {
		channelID := 0
		fmt.Sscanf(id, "%d", &channelID)
		epgData := getEPG(channelID)
		channel = map[string]interface{}{
			"ID":                id,
			"Name":              name.String,
			"StreamURL":         streamUrl.String,
			"LogoURL":           logoUrl.String,
			"YoutubeURL":        youtubeUrl.String,
			"YoutubeChannelURL": youtubeChannelUrl.String,
			"EPGData":           epgData,
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
	youtubeChannelUrl := r.FormValue("youtube_channel_url")

	_, err := config.DB.Exec(`UPDATE live_tv_channels SET name=?, stream_url=NULLIF(?,''), logo_url=NULLIF(?,''), youtube_url=NULLIF(?,''), youtube_channel_url=NULLIF(?,'') WHERE id=?`, name, streamUrl, logoUrl, youtubeUrl, youtubeChannelUrl, id)
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

func AdminSettingsView(w http.ResponseWriter, r *http.Request) {
	var upNextTimer, frontendMenu string
	config.DB.QueryRow("SELECT setting_value FROM app_settings WHERE setting_key = 'up_next_timer'").Scan(&upNextTimer)
	config.DB.QueryRow("SELECT setting_value FROM app_settings WHERE setting_key = 'frontend_menu'").Scan(&frontendMenu)
	
	// Remove surrounding quotes for timer if present
	if len(upNextTimer) >= 2 && upNextTimer[0] == '"' && upNextTimer[len(upNextTimer)-1] == '"' {
		upNextTimer = upNextTimer[1 : len(upNextTimer)-1]
	}

	data := map[string]interface{}{
		"UpNextTimer": upNextTimer,
		"FrontendMenu": frontendMenu,
	}
	renderTemplate(w, "admin_settings.html", data)
}

func AdminSettingsUpdate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	timer := r.FormValue("up_next_timer")
	menu := r.FormValue("frontend_menu")
	
	config.DB.Exec("UPDATE app_settings SET setting_value = ? WHERE setting_key = 'up_next_timer'", fmt.Sprintf(`"%s"`, timer))
	config.DB.Exec("UPDATE app_settings SET setting_value = ? WHERE setting_key = 'frontend_menu'", menu)
	
	http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
}

func AdminMusicView(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT id, title, artist FROM music_content ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var music []map[string]interface{}
	for rows.Next() {
		var id int
		var title string
		var artist sql.NullString
		if err := rows.Scan(&id, &title, &artist); err == nil {
			music = append(music, map[string]interface{}{
				"ID":     id,
				"Title":  title,
				"Artist": artist.String,
			})
		}
	}
	renderTemplate(w, "admin_music.html", map[string]interface{}{"Music": music})
}

func AdminMusicFormView(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin_music_form.html", nil)
}

func AdminMusicCreate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	title := r.FormValue("title")
	artist := r.FormValue("artist")
	desc := r.FormValue("description")
	releaseDate := r.FormValue("release_date")
	posterUrl := r.FormValue("poster_url")
	backdropUrl := r.FormValue("backdrop_url")
	videoUrl := r.FormValue("video_url")
	
	vidSrc := "[]"
	if videoUrl != "" {
		vidSrc = fmt.Sprintf(`[{"quality": "Original", "url": "%s"}]`, videoUrl)
	}

	_, err := config.DB.Exec(`INSERT INTO music_content (title, description, artist, release_date, poster_url, backdrop_url, video_sources) 
		VALUES (?, NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), ?)`, 
		title, desc, artist, releaseDate, posterUrl, backdropUrl, vidSrc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/music", http.StatusSeeOther)
}

func AdminMusicDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id != "" {
		config.DB.Exec("DELETE FROM music_content WHERE id = ?", id)
	}
	http.Redirect(w, r, "/admin/music", http.StatusSeeOther)
}

func AdminCategoriesView(w http.ResponseWriter, r *http.Request) {
    types := []string{"movie", "series", "live_tv", "sports", "music"}
    categoriesByType := make(map[string][]map[string]interface{})

    for _, t := range types {
        tableName := t + "_categories"
        rows, err := config.DB.Query(fmt.Sprintf("SELECT id, name, slug FROM %s ORDER BY name ASC", tableName))
        if err == nil {
            var cats []map[string]interface{}
            for rows.Next() {
                var id int
                var name, slug string
                rows.Scan(&id, &name, &slug)
                cats = append(cats, map[string]interface{}{
                    "ID": id,
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

func AdminCategoryCreate(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    t := r.FormValue("type")
    name := r.FormValue("name")
    
    if t != "" && name != "" {
        slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
        tableName := t + "_categories"
        config.DB.Exec(fmt.Sprintf("INSERT IGNORE INTO %s (name, slug) VALUES (?, ?)", tableName), name, slug)
    }
    http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}

func AdminCategoryDelete(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    t := r.FormValue("type")
    id := r.FormValue("id")
    
    if t != "" && id != "" {
        tableName := t + "_categories"
        config.DB.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName), id)
    }
    http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}

func AdminFetchYouTubeLive(w http.ResponseWriter, r *http.Request) {
	channelInput := r.URL.Query().Get("channel")
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		http.Error(w, "YOUTUBE_API_KEY not configured in environment variables.", http.StatusInternalServerError)
		return
	}

	channelID := ""

	// Parse input
	if strings.Contains(channelInput, "channel/") {
		parts := strings.Split(channelInput, "channel/")
		if len(parts) > 1 {
			channelID = strings.Split(parts[1], "/")[0]
			channelID = strings.Split(channelID, "?")[0]
		}
	} else if strings.Contains(channelInput, "@") {
		parts := strings.Split(channelInput, "@")
		if len(parts) > 1 {
			handle := "@" + strings.Split(parts[1], "/")[0]
			handle = strings.Split(handle, "?")[0]
			
			// Lookup channel ID from handle
			apiURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/channels?part=id&forHandle=%s&key=%s", url.QueryEscape(handle), apiKey)
			resp, err := http.Get(apiURL)
			if err == nil {
				defer resp.Body.Close()
				bodyBytes, _ := io.ReadAll(resp.Body)
				var data struct {
					Items []struct {
						Id string `json:"id"`
					} `json:"items"`
				}
				json.Unmarshal(bodyBytes, &data)
				if len(data.Items) > 0 {
					channelID = data.Items[0].Id
				} else {
					log.Printf("YouTube API (forHandle) returned no items: %s", string(bodyBytes))
					// Fallback to forUsername
					username := strings.TrimPrefix(handle, "@")
					apiURL = fmt.Sprintf("https://www.googleapis.com/youtube/v3/channels?part=id&forUsername=%s&key=%s", url.QueryEscape(username), apiKey)
					resp2, err2 := http.Get(apiURL)
					if err2 == nil {
						defer resp2.Body.Close()
						bodyBytes2, _ := io.ReadAll(resp2.Body)
						json.Unmarshal(bodyBytes2, &data)
						if len(data.Items) > 0 {
							channelID = data.Items[0].Id
						} else {
							log.Printf("YouTube API (forUsername) returned no items: %s", string(bodyBytes2))
						}
					}
				}
			}
		}
	} else if len(channelInput) == 24 && strings.HasPrefix(channelInput, "UC") {
        // Direct channel ID
        channelID = channelInput
    }

	if channelID == "" {
		http.Error(w, "Could not extract channel ID. Please provide a valid YouTube channel URL or @handle.", http.StatusBadRequest)
		return
	}

	// Search for live video
	searchUrl := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?part=id&channelId=%s&eventType=live&type=video&key=%s", channelID, apiKey)
	resp, err := http.Get(searchUrl)
	if err != nil {
		http.Error(w, "Failed to query YouTube API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
        http.Error(w, fmt.Sprintf("YouTube API returned status %d. Please check your API key.", resp.StatusCode), http.StatusInternalServerError)
		return
    }

	var searchData struct {
		Items []struct {
			Id struct {
				VideoId string `json:"videoId"`
			} `json:"id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&searchData); err != nil {
		http.Error(w, "Failed to parse YouTube API response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(searchData.Items) == 0 {
		json.NewEncoder(w).Encode(map[string]string{"error": "This channel does not currently have an active live stream."})
		return
	}

	videoUrl := "https://www.youtube.com/watch?v=" + searchData.Items[0].Id.VideoId
	json.NewEncoder(w).Encode(map[string]string{"live_url": videoUrl})
}

func AdminFetchAllYouTubeLive(w http.ResponseWriter, r *http.Request) {
	// First fetch all channels with a youtube_channel_url
	rows, err := config.DB.Query("SELECT id, youtube_channel_url FROM live_tv_channels WHERE youtube_channel_url IS NOT NULL AND youtube_channel_url != ''")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ChannelInfo struct {
		ID  int
		URL string
	}
	var channels []ChannelInfo
	for rows.Next() {
		var id int
		var url string
		if err := rows.Scan(&id, &url); err == nil {
			channels = append(channels, ChannelInfo{ID: id, URL: url})
		}
	}
	rows.Close()

	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		http.Error(w, "YOUTUBE_API_KEY not configured", http.StatusInternalServerError)
		return
	}

	updatedCount := 0
	for _, ch := range channels {
		channelInput := ch.URL
		channelID := ""

		if strings.Contains(channelInput, "channel/") {
			parts := strings.Split(channelInput, "channel/")
			if len(parts) > 1 {
				channelID = strings.Split(parts[1], "/")[0]
				channelID = strings.Split(channelID, "?")[0]
			}
		} else if strings.Contains(channelInput, "@") {
			parts := strings.Split(channelInput, "@")
			if len(parts) > 1 {
				handle := "@" + strings.Split(parts[1], "/")[0]
				handle = strings.Split(handle, "?")[0]
				apiURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/channels?part=id&forHandle=%s&key=%s", url.QueryEscape(handle), apiKey)
				resp, err := http.Get(apiURL)
				if err == nil {
					bodyBytes, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					var data struct {
						Items []struct {
							Id string `json:"id"`
						} `json:"items"`
					}
					json.Unmarshal(bodyBytes, &data)
					if len(data.Items) > 0 {
						channelID = data.Items[0].Id
					} else {
						username := strings.TrimPrefix(handle, "@")
						apiURL = fmt.Sprintf("https://www.googleapis.com/youtube/v3/channels?part=id&forUsername=%s&key=%s", url.QueryEscape(username), apiKey)
						resp2, err2 := http.Get(apiURL)
						if err2 == nil {
							bodyBytes2, _ := io.ReadAll(resp2.Body)
							resp2.Body.Close()
							json.Unmarshal(bodyBytes2, &data)
							if len(data.Items) > 0 {
								channelID = data.Items[0].Id
							}
						}
					}
				}
			}
		} else if len(channelInput) == 24 && strings.HasPrefix(channelInput, "UC") {
			channelID = channelInput
		}

		if channelID != "" {
			searchUrl := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?part=id&channelId=%s&eventType=live&type=video&key=%s", channelID, apiKey)
			resp, err := http.Get(searchUrl)
			if err == nil && resp.StatusCode == 200 {
				var searchData struct {
					Items []struct {
						Id struct {
							VideoId string `json:"videoId"`
						} `json:"id"`
					} `json:"items"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&searchData); err == nil {
					if len(searchData.Items) > 0 {
						videoUrl := "https://www.youtube.com/watch?v=" + searchData.Items[0].Id.VideoId
						config.DB.Exec("UPDATE live_tv_channels SET youtube_url=? WHERE id=?", videoUrl, ch.ID)
						updatedCount++
					}
				}
				resp.Body.Close()
			} else if resp != nil {
				resp.Body.Close()
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"updated_count": updatedCount,
	})
}
