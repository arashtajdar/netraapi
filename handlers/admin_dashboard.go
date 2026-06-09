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

	// Mocking DAU / Active Viewers for the dashboard demo
	stats.ActiveViewers = 42
	stats.DAUChartLabels = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	stats.DAUChartData = []int{120, 150, 180, 130, 200, 250, 310}

	// Real Popular Content
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

	if len(stats.PopularTitles) == 0 {
		stats.PopularTitles = []string{"Inception", "Interstellar", "Dark Knight"}
		stats.PopularCounts = []int{45, 30, 25}
	}

	renderTemplate(w, "admin_dashboard.html", stats)
}
