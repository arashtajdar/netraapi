package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"sheedbox-api/config"
	"sheedbox-api/models"
	"sheedbox-api/services"

	"github.com/go-chi/chi/v5"
)

type EPGData struct {
	ProgramTitle string `json:"program_title"`
	Description  string `json:"description"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
}

type AdminLiveTVHandler struct {
	livetvService *services.LiveTVService
}

func NewAdminLiveTVHandler(livetvService *services.LiveTVService) *AdminLiveTVHandler {
	return &AdminLiveTVHandler{livetvService: livetvService}
}

func (h *AdminLiveTVHandler) View(w http.ResponseWriter, r *http.Request) {
	channels, err := h.livetvService.ListChannels(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var channelData []map[string]interface{}
	for _, c := range channels {
		var streamURL, logoURL, youtubeURL, youtubeChannelURL string
		if c.StreamURL != nil {
			streamURL = *c.StreamURL
		}
		if c.LogoURL != nil {
			logoURL = *c.LogoURL
		}
		if c.YoutubeURL != nil {
			youtubeURL = *c.YoutubeURL
		}
		if c.YoutubeChannelURL != nil {
			youtubeChannelURL = *c.YoutubeChannelURL
		}
		channelData = append(channelData, map[string]interface{}{
			"ID":                c.ID,
			"Name":              c.Name,
			"Slug":              c.Slug,
			"StreamURL":         streamURL,
			"LogoURL":           logoURL,
			"YoutubeURL":        youtubeURL,
			"YoutubeChannelURL": youtubeChannelURL,
			"EPGFetchURL":       c.EPGFetchURL,
			"LastEPGFetch":      c.LastEPGFetch,
			"NextEPGFetch":      c.NextEPGFetch,
		})
	}

	renderTemplate(w, "admin_livetv.html", map[string]interface{}{"Channels": channelData})
}

func (h *AdminLiveTVHandler) FormView(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "admin_livetv_form.html", nil)
}

func (h *AdminLiveTVHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.FormValue("name")
	slug := r.FormValue("slug")
	streamUrl := r.FormValue("stream_url")
	logoUrl := r.FormValue("logo_url")
	youtubeUrl := r.FormValue("youtube_url")
	youtubeChannelUrl := r.FormValue("youtube_channel_url")
	epgFetchUrl := r.FormValue("epg_fetch_url")

	channel := models.LiveTVChannel{
		Name:              name,
		Slug:              slug,
		StreamURL:         stringPtr(streamUrl),
		LogoURL:           stringPtr(logoUrl),
		YoutubeURL:        stringPtr(youtubeUrl),
		YoutubeChannelURL: stringPtr(youtubeChannelUrl),
		EPGFetchURL:       stringPtr(epgFetchUrl),
	}

	channelID, err := h.livetvService.CreateChannel(r.Context(), &channel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.saveEPG(r.Context(), channelID, r.FormValue("epg_data"))

	config.ClearCachePattern("livetv_*")

	http.Redirect(w, r, "/admin/live-tv", http.StatusSeeOther)
}

func (h *AdminLiveTVHandler) EditFormView(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	channel, err := h.livetvService.GetChannel(r.Context(), id)
	if err != nil || channel == nil {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	epgList, err := h.livetvService.GetEPG(r.Context(), id)
	if err != nil {
		epgList = []models.EPG{}
	}

	epgJSON := "[]"
	if len(epgList) > 0 {
		var items []EPGData
		for _, e := range epgList {
			desc := ""
			if e.Description != nil {
				desc = *e.Description
			}
			items = append(items, EPGData{
				ProgramTitle: e.ProgramTitle,
				Description:  desc,
				StartTime:    e.StartTime.Format(time.RFC3339),
				EndTime:      e.EndTime.Format(time.RFC3339),
			})
		}
		if b, err := json.MarshalIndent(items, "", "  "); err == nil {
			epgJSON = string(b)
		}
	}

	var streamURL, logoURL, youtubeURL, youtubeChannelURL string
	if channel.StreamURL != nil {
		streamURL = *channel.StreamURL
	}
	if channel.LogoURL != nil {
		logoURL = *channel.LogoURL
	}
	if channel.YoutubeURL != nil {
		youtubeURL = *channel.YoutubeURL
	}
	if channel.YoutubeChannelURL != nil {
		youtubeChannelURL = *channel.YoutubeChannelURL
	}

	channelMap := map[string]interface{}{
		"ID":                channel.ID,
		"Name":              channel.Name,
		"Slug":              channel.Slug,
		"StreamURL":         streamURL,
		"LogoURL":           logoURL,
		"YoutubeURL":        youtubeURL,
		"YoutubeChannelURL": youtubeChannelURL,
		"EPGFetchURL":       channel.EPGFetchURL,
		"LastEPGFetch":      channel.LastEPGFetch,
		"NextEPGFetch":      channel.NextEPGFetch,
		"EPGData":           epgJSON,
	}

	renderTemplate(w, "admin_livetv_form.html", map[string]interface{}{"Channel": channelMap})
}

func (h *AdminLiveTVHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	r.ParseForm()
	name := r.FormValue("name")
	slug := r.FormValue("slug")
	streamUrl := r.FormValue("stream_url")
	logoUrl := r.FormValue("logo_url")
	youtubeUrl := r.FormValue("youtube_url")
	youtubeChannelUrl := r.FormValue("youtube_channel_url")
	epgFetchUrl := r.FormValue("epg_fetch_url")

	channel := models.LiveTVChannel{
		ID:                id,
		Name:              name,
		Slug:              slug,
		StreamURL:         stringPtr(streamUrl),
		LogoURL:           stringPtr(logoUrl),
		YoutubeURL:        stringPtr(youtubeUrl),
		YoutubeChannelURL: stringPtr(youtubeChannelUrl),
		EPGFetchURL:       stringPtr(epgFetchUrl),
		LastEPGFetch:      nil, // preserve via service if needed, but since we use update, wait! 
	}
	
	// Better to retrieve the existing channel to preserve LastEPGFetch and NextEPGFetch
	existing, _ := h.livetvService.GetChannel(r.Context(), id)
	if existing != nil {
		channel.LastEPGFetch = existing.LastEPGFetch
		channel.NextEPGFetch = existing.NextEPGFetch
	}

	err = h.livetvService.UpdateChannel(r.Context(), &channel)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.saveEPG(r.Context(), int64(id), r.FormValue("epg_data"))

	config.ClearCachePattern("livetv_*")

	http.Redirect(w, r, "/admin/live-tv", http.StatusSeeOther)
}

func (h *AdminLiveTVHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || idStr == "" {
		http.Error(w, "Missing or invalid channel ID", http.StatusBadRequest)
		return
	}

	err = h.livetvService.DeleteChannel(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to delete channel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	config.ClearCachePattern("livetv_*")

	http.Redirect(w, r, "/admin/live-tv", http.StatusSeeOther)
}

func (h *AdminLiveTVHandler) saveEPG(ctx context.Context, channelID int64, epgJSON string) {
	if epgJSON == "" || epgJSON == "[]" {
		h.livetvService.SaveEPG(ctx, channelID, nil)
		return
	}

	var parsed []EPGData
	err := json.Unmarshal([]byte(epgJSON), &parsed)
	if err != nil {
		log.Println("Invalid EPG JSON:", err)
		return
	}

	var epgs []models.EPG
	for _, item := range parsed {
		var start, end time.Time
		if t, err := time.Parse(time.RFC3339, item.StartTime); err == nil {
			start = t
		} else if t, err := time.Parse("2006-01-02 15:04:05", item.StartTime); err == nil {
			start = t
		}
		if t, err := time.Parse(time.RFC3339, item.EndTime); err == nil {
			end = t
		} else if t, err := time.Parse("2006-01-02 15:04:05", item.EndTime); err == nil {
			end = t
		}

		epgs = append(epgs, models.EPG{
			ChannelID:    int(channelID),
			ProgramTitle: item.ProgramTitle,
			Description:  stringPtr(item.Description),
			StartTime:    start,
			EndTime:      end,
		})
	}

	h.livetvService.SaveEPG(ctx, channelID, epgs)
}

func (h *AdminLiveTVHandler) FetchYouTubeLive(w http.ResponseWriter, r *http.Request) {
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
		channelID = channelInput
	}

	if channelID == "" {
		http.Error(w, "Could not extract channel ID. Please provide a valid YouTube channel URL or @handle.", http.StatusBadRequest)
		return
	}

	// Search for live video
	searchUrl := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?part=id,snippet&channelId=%s&eventType=live&type=video&key=%s", channelID, apiKey)
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
			Snippet struct {
				Title string `json:"title"`
			} `json:"snippet"`
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

	type StreamData struct {
		VideoID string `json:"video_id"`
		Title   string `json:"title"`
		URL     string `json:"url"`
	}
	var streams []StreamData

	for _, item := range searchData.Items {
		streams = append(streams, StreamData{
			VideoID: item.Id.VideoId,
			Title:   item.Snippet.Title,
			URL:     "https://www.youtube.com/watch?v=" + item.Id.VideoId,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"streams": streams})
}

func (h *AdminLiveTVHandler) FetchAllYouTubeLive(w http.ResponseWriter, r *http.Request) {
	channels, err := h.livetvService.ListChannels(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		http.Error(w, "YOUTUBE_API_KEY not configured", http.StatusInternalServerError)
		return
	}

	updatedCount := 0
	for _, ch := range channels {
		if ch.YoutubeChannelURL == nil || *ch.YoutubeChannelURL == "" {
			continue
		}

		channelInput := *ch.YoutubeChannelURL
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
						h.livetvService.UpdateYoutubeURL(r.Context(), ch.ID, videoUrl)
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
		"success":       true,
		"updated_count": updatedCount,
	})
}

type RemoteEPGItem struct {
	Title     string `json:"title"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

func (h *AdminLiveTVHandler) FetchEPGData(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	channel, err := h.livetvService.GetChannel(r.Context(), id)
	if err != nil || channel == nil {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	if channel.EPGFetchURL == nil || *channel.EPGFetchURL == "" {
		http.Error(w, "No EPG Fetch URL defined for this channel", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(*channel.EPGFetchURL)
	if err != nil {
		http.Error(w, "Failed to fetch from EPG URL: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var remoteData []RemoteEPGItem
	if err := json.NewDecoder(resp.Body).Decode(&remoteData); err != nil {
		http.Error(w, "Failed to parse EPG JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Filter and convert remote data
	now := time.Now().UTC()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	var newEPGs []models.EPG

	for _, item := range remoteData {
		var start, end time.Time
		if t, err := time.Parse(time.RFC3339, item.StartTime); err == nil {
			start = t
		} else if t, err := time.Parse("2006-01-02 15:04:05", item.StartTime); err == nil {
			start = t
		}
		if t, err := time.Parse(time.RFC3339, item.EndTime); err == nil {
			end = t
		} else if t, err := time.Parse("2006-01-02 15:04:05", item.EndTime); err == nil {
			end = t
		}

		// Only keep from today onwards
		if end.Before(startOfToday) {
			continue
		}

		newEPGs = append(newEPGs, models.EPG{
			ChannelID:    id,
			ProgramTitle: item.Title,
			StartTime:    start,
			EndTime:      end,
		})
	}

	// Fetch existing EPG to keep past ones if they wanted (or replace all if they wanted only today onwards).
	// "we only want the data from today and so on so ignore all data from yesteday and previous days"
	// This means we can just save only the newEPGs and it effectively clears out the old data.
	// But let's also keep existing ones from DB that are before today? No, the prompt says "ignore all data from yesterday and previous days", so discarding them is likely what they mean to keep the table small.
	
	err = h.livetvService.SaveEPG(r.Context(), int64(id), newEPGs)
	if err != nil {
		http.Error(w, "Failed to save EPG: "+err.Error(), http.StatusInternalServerError)
		return
	}

	nowTime := time.Now()
	nextTime := nowTime.Add(24 * time.Hour)
	channel.LastEPGFetch = &nowTime
	channel.NextEPGFetch = &nextTime
	h.livetvService.UpdateChannel(r.Context(), channel)

	config.ClearCachePattern("livetv_*")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"count":   len(newEPGs),
	})
}
