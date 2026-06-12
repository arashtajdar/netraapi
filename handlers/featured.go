package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"sheedbox-api/config"
	"sheedbox-api/contextkeys"
)

type FeaturedItem struct {
	ID                int    `json:"id"`
	ContentType       string `json:"content_type"`
	ContentID         int    `json:"content_id"`
	CustomDescription string `json:"custom_description"`
	Title             string `json:"title"`
	PosterURL         string `json:"poster_url"`
	BackdropURL       string `json:"backdrop_url"`
}

type FeaturedHandler struct {
}

func NewFeaturedHandler() *FeaturedHandler {
	return &FeaturedHandler{}
}

func (h *FeaturedHandler) GetFeatured(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userLevel := contextkeys.UserLevelFromContext(r.Context())
	cacheKey := fmt.Sprintf("featured_items:level:%d", userLevel)

	// Check Redis Cache
	if config.RedisClient != nil {
		cached, err := config.RedisClient.Get(context.Background(), cacheKey).Result()
		if err == nil {
			w.Write([]byte(cached))
			return
		}
	}

	// Check if there are any featured items in the database
	rows, err := config.DB.Query(`
		SELECT f.id, f.content_type, f.content_id, f.custom_description,
		CASE f.content_type
			WHEN 'movie' THEN (SELECT title FROM movies WHERE id = f.content_id)
			WHEN 'series' THEN (SELECT title FROM series WHERE id = f.content_id)
			WHEN 'live_tv' THEN (SELECT name FROM live_tv_channels WHERE id = f.content_id)
			WHEN 'sports' THEN (SELECT title FROM sports_events WHERE id = f.content_id)
		END as content_title,
		COALESCE(NULLIF(f.image_url, ''), CASE f.content_type
			WHEN 'movie' THEN (SELECT poster_url FROM movies WHERE id = f.content_id)
			WHEN 'series' THEN (SELECT poster_url FROM series WHERE id = f.content_id)
			WHEN 'live_tv' THEN (SELECT logo_url FROM live_tv_channels WHERE id = f.content_id)
			WHEN 'sports' THEN NULL
		END) as poster_url,
		COALESCE(NULLIF(f.image_url, ''), CASE f.content_type
			WHEN 'movie' THEN (SELECT backdrop_url FROM movies WHERE id = f.content_id)
			WHEN 'series' THEN (SELECT backdrop_url FROM series WHERE id = f.content_id)
			WHEN 'live_tv' THEN (SELECT logo_url FROM live_tv_channels WHERE id = f.content_id)
			WHEN 'sports' THEN NULL
		END) as backdrop_url
		FROM featured_items f
		WHERE (CASE f.content_type
			WHEN 'movie' THEN (SELECT access_level FROM movies WHERE id = f.content_id)
			WHEN 'series' THEN (SELECT access_level FROM series WHERE id = f.content_id)
			WHEN 'live_tv' THEN (SELECT access_level FROM live_tv_channels WHERE id = f.content_id)
			WHEN 'sports' THEN (SELECT access_level FROM sports_events WHERE id = f.content_id)
			ELSE 1
		END) <= ?
		ORDER BY f.created_at DESC
	`, userLevel)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var featured []FeaturedItem
	for rows.Next() {
		var item FeaturedItem
		var customDesc, title, poster, backdrop *string
		if err := rows.Scan(&item.ID, &item.ContentType, &item.ContentID, &customDesc, &title, &poster, &backdrop); err == nil {
			if customDesc != nil { item.CustomDescription = *customDesc }
			if title != nil { item.Title = *title }
			if poster != nil { item.PosterURL = *poster }
			if backdrop != nil { item.BackdropURL = *backdrop }
			featured = append(featured, item)
		}
	}

	// Fallback to random items if no featured items exist
	if len(featured) == 0 {
		fallbackRows, err := config.DB.Query(`
			(SELECT id as content_id, 'movie' as content_type, title, poster_url, backdrop_url, description FROM movies WHERE access_level <= ? ORDER BY RAND() LIMIT 5)
			UNION ALL
			(SELECT id as content_id, 'series' as content_type, title, poster_url, backdrop_url, description FROM series WHERE access_level <= ? ORDER BY RAND() LIMIT 5)
			ORDER BY RAND() LIMIT 5
		`, userLevel, userLevel)
		if err == nil {
			defer fallbackRows.Close()
			for fallbackRows.Next() {
				var item FeaturedItem
				var title, poster, backdrop, desc *string
				if err := fallbackRows.Scan(&item.ContentID, &item.ContentType, &title, &poster, &backdrop, &desc); err == nil {
					if title != nil { item.Title = *title }
					if poster != nil { item.PosterURL = *poster }
					if backdrop != nil { item.BackdropURL = *backdrop }
					if desc != nil { item.CustomDescription = *desc }
					featured = append(featured, item)
				}
			}
		}
	}

	if featured == nil {
		featured = []FeaturedItem{}
	}

	respBytes, _ := json.Marshal(featured)
	
	// Save to Redis Cache (expires in 10 minutes)
	if config.RedisClient != nil {
		config.RedisClient.Set(context.Background(), cacheKey, respBytes, 10*time.Minute)
	}

	w.Write(respBytes)
}
