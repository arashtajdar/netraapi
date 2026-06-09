package handlers

import (
	"database/sql"
	"net/http"
)

type DashboardStats struct {
	MoviesCount    int
	SeriesCount    int
	LiveTVCount    int
	SportsCount    int
	TotalCoins     int
	DAUChartLabels []string
	DAUChartData   []int
	PopularTitles  []string
	PopularCounts  []int
	ActiveViewers  int
}

type AdminDashboardHandler struct {
	db *sql.DB
}

func NewAdminDashboardHandler(db *sql.DB) *AdminDashboardHandler {
	return &AdminDashboardHandler{db: db}
}

func (h *AdminDashboardHandler) View(w http.ResponseWriter, r *http.Request) {
	var stats DashboardStats

	h.db.QueryRow("SELECT COUNT(*) FROM movies").Scan(&stats.MoviesCount)
	h.db.QueryRow("SELECT COUNT(*) FROM series").Scan(&stats.SeriesCount)
	h.db.QueryRow("SELECT COUNT(*) FROM live_tv_channels").Scan(&stats.LiveTVCount)
	h.db.QueryRow("SELECT COUNT(*) FROM sports_events").Scan(&stats.SportsCount)
	h.db.QueryRow("SELECT COALESCE(SUM(virtual_coins), 0) FROM users").Scan(&stats.TotalCoins)

	// Real Active Viewers (Active in last 15 minutes)
	h.db.QueryRow(`
		SELECT COUNT(DISTINCT user_id) 
		FROM user_watch_history 
		WHERE updated_at >= NOW() - INTERVAL 15 MINUTE
	`).Scan(&stats.ActiveViewers)

	// Real DAU (Daily Active Users over the last 7 days)
	// We'll initialize empty slices
	stats.DAUChartLabels = []string{}
	stats.DAUChartData = []int{}

	dauRows, err := h.db.Query(`
		SELECT DATE_FORMAT(updated_at, '%a'), COUNT(DISTINCT user_id) 
		FROM user_watch_history 
		WHERE updated_at >= DATE_SUB(CURDATE(), INTERVAL 6 DAY)
		GROUP BY DATE(updated_at), DATE_FORMAT(updated_at, '%a')
		ORDER BY DATE(updated_at) ASC
	`)
	if err == nil {
		defer dauRows.Close()
		for dauRows.Next() {
			var day string
			var count int
			if err := dauRows.Scan(&day, &count); err == nil {
				stats.DAUChartLabels = append(stats.DAUChartLabels, day)
				stats.DAUChartData = append(stats.DAUChartData, count)
			}
		}
	}

	// Real Popular Content
	stats.PopularTitles = []string{}
	stats.PopularCounts = []int{}
	
	rows, err := h.db.Query(`
		SELECT m.title, COUNT(uwh.movie_id) as views 
		FROM user_watch_history uwh 
		JOIN movies m ON uwh.movie_id = m.id 
		GROUP BY uwh.movie_id 
		ORDER BY views DESC LIMIT 5
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var title string
			var count int
			if err := rows.Scan(&title, &count); err == nil {
				stats.PopularTitles = append(stats.PopularTitles, title)
				stats.PopularCounts = append(stats.PopularCounts, count)
			}
		}
	}

	renderTemplate(w, "admin_dashboard.html", stats)
}
