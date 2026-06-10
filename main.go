package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"sheedbox-api/config"
	"sheedbox-api/handlers"
	"sheedbox-api/middleware"
	"sheedbox-api/repository/mysql"
	"sheedbox-api/services"
	"sheedbox-api/services/storage"
	"sheedbox-api/websockets"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Initialize structured logging (slog)
	var logHandler slog.Handler
	if os.Getenv("ENV") == "production" || os.Getenv("RAILWAY_ENVIRONMENT") != "" {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	slog.SetDefault(slog.New(logHandler))

	if err := godotenv.Load(); err != nil {
		slog.Warn("Warning: No .env file found or error reading it")
	}

	// Initialize JWT — panics if JWT_SECRET is missing (fail-fast security)
	config.InitJWT()

	config.ConnectDB()
	defer config.DB.Close()
	config.ConnectRedis()

	// Pre-cache HTML templates for production admin page rendering
	handlers.InitTemplates()

	// Initialize admin authentication
	middleware.InitAdminAuth()

	if os.Getenv("STORAGE_DRIVER") == "r2" {
		storage.ActiveProvider = storage.NewR2Storage(
			os.Getenv("R2_ACCOUNT_ID"),
			os.Getenv("R2_ACCESS_KEY_ID"),
			os.Getenv("R2_SECRET_ACCESS_KEY"),
			os.Getenv("R2_BUCKET_NAME"),
			os.Getenv("R2_PUBLIC_URL"),
		)
		slog.Info("☁️ Cloudflare R2 Storage Provider Initialized")
	} else {
		storage.ActiveProvider = storage.NewLocalStorage(
			os.Getenv("STORAGE_LOCAL_DIR"),
			os.Getenv("STORAGE_LOCAL_URL"),
		)
		slog.Info("📂 Local Storage Provider Initialized")
	}

	hub := websockets.NewHub()
	go hub.Run()

	services.InitVideoProcessor()

	r := chi.NewRouter()

	r.Use(middleware.RequestLogger)

	allowedOrigins := []string{"https://*", "http://*"}
	if originsEnv := os.Getenv("ALLOWED_ORIGINS"); originsEnv != "" {
		allowedOrigins = strings.Split(originsEnv, ",")
		for i := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
		}
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Profile-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Liveness and Readiness Probes for deploys
	r.Get("/healthz", handlers.LivenessCheck)
	r.Get("/readyz", handlers.ReadinessCheck)

	// Initialize DI for Movies domain
	movieRepo := mysql.NewMovieRepository(config.DB)
	movieService := services.NewMovieService(movieRepo)
	movieHandler := handlers.NewMovieHandler(movieService)

	// Initialize DI for Series domain
	seriesRepo := mysql.NewSeriesRepository(config.DB)
	seriesService := services.NewSeriesService(seriesRepo)
	seriesHandler := handlers.NewSeriesHandler(seriesService)

	// Initialize DI for Live TV domain
	livetvRepo := mysql.NewLiveTVRepository(config.DB)
	livetvService := services.NewLiveTVService(livetvRepo)
	livetvHandler := handlers.NewLiveTVHandler(livetvService)

	// Initialize DI for Sports domain
	sportsRepo := mysql.NewSportsRepository(config.DB)
	sportsService := services.NewSportsService(sportsRepo)
	sportsHandler := handlers.NewSportsHandler(sportsService)

	// Initialize DI for Music domain
	musicRepo := mysql.NewMusicRepository(config.DB)
	musicService := services.NewMusicService(musicRepo)
	musicHandler := handlers.NewMusicHandler(musicService)

	// Initialize DI for Featured domain
	featuredHandler := handlers.NewFeaturedHandler()

	// Initialize DI for User & Auth domain
	userRepo := mysql.NewUserRepository(config.DB)
	userService := services.NewUserService(userRepo)
	authHandler := handlers.NewAuthHandler(userService)
	gamificationHandler := handlers.NewGamificationHandler(userService)

	// Initialize DI for Profiles domain
	profileRepo := mysql.NewUserProfileRepository(config.DB)
	profileService := services.NewProfileService(profileRepo)
	profileHandler := handlers.NewProfileHandler(profileService)

	// Initialize DI for Watchlists domain
	watchlistRepo := mysql.NewWatchlistRepository(config.DB)
	watchlistService := services.NewWatchlistService(watchlistRepo)
	watchlistHandler := handlers.NewWatchlistHandler(watchlistService)

	// Initialize DI for Settings, UpNext, Recommendations
	settingsHandler := handlers.NewSettingsHandler(config.DB)
	upNextHandler := handlers.NewUpNextHandler(config.DB)
	recommendationHandler := handlers.NewRecommendationHandler(config.DB)

	// Initialize DI for Admin handlers
	adminDashboardHandler := handlers.NewAdminDashboardHandler(config.DB)
	adminMovieHandler := handlers.NewAdminMovieHandler(movieService)
	adminSeriesHandler := handlers.NewAdminSeriesHandler(seriesService)
	adminLiveTVHandler := handlers.NewAdminLiveTVHandler(livetvService)
	adminSportsHandler := handlers.NewAdminSportsHandler(sportsService)
	adminMusicHandler := handlers.NewAdminMusicHandler(musicService)
	adminSettingsHandler := handlers.NewAdminSettingsHandler(config.DB)
	adminCategoriesHandler := handlers.NewAdminCategoriesHandler(config.DB)
	adminUsersHandler := handlers.NewAdminUsersHandler(config.DB)
	adminFeaturedHandler := handlers.NewAdminFeaturedHandler(config.DB)

	// Versioned and Backwards-Compatible Routing Setup
	r.Route("/api", func(r chi.Router) {
		// Mirror directly on /api for backwards compatibility with the existing client
		registerAPIRoutes(r, authHandler, gamificationHandler, profileHandler, watchlistHandler, settingsHandler, upNextHandler, featuredHandler, movieHandler, seriesHandler, livetvHandler, sportsHandler, musicHandler, recommendationHandler)

		// API V1 endpoint group
		r.Route("/v1", func(r chi.Router) {
			registerAPIRoutes(r, authHandler, gamificationHandler, profileHandler, watchlistHandler, settingsHandler, upNextHandler, featuredHandler, movieHandler, seriesHandler, livetvHandler, sportsHandler, musicHandler, recommendationHandler)
		})
	})

	// Admin UI Routes — protected by admin session authentication
	r.Route("/admin", func(r chi.Router) {
		// Login/logout routes are OUTSIDE the auth middleware
		r.Get("/login", middleware.AdminLoginView)
		r.Post("/login", middleware.AdminLoginSubmit)
		r.Get("/logout", middleware.AdminLogout)

		// All other admin routes require authentication
		r.Group(func(r chi.Router) {
			r.Use(middleware.AdminAuthMiddleware)

			r.Get("/", adminDashboardHandler.View)
			r.Post("/upload", handlers.UploadMedia)
			r.Get("/movies", adminMovieHandler.View)
			r.Get("/movies/new", adminMovieHandler.FormView)
			r.Post("/movies/new", adminMovieHandler.Create)
			r.Get("/series", adminSeriesHandler.View)
			r.Get("/series/new", adminSeriesHandler.FormView)
			r.Post("/series/new", adminSeriesHandler.Create)
			r.Get("/live-tv", adminLiveTVHandler.View)
			r.Get("/live-tv/new", adminLiveTVHandler.FormView)
			r.Post("/live-tv/new", adminLiveTVHandler.Create)
			r.Get("/live-tv/edit/{id}", adminLiveTVHandler.EditFormView)
			r.Post("/live-tv/edit/{id}", adminLiveTVHandler.Update)
			r.Post("/live-tv/delete/{id}", adminLiveTVHandler.Delete)
			r.Post("/live-tv/fetch-epg/{id}", adminLiveTVHandler.FetchEPGData)
			r.Get("/sports", adminSportsHandler.View)
			r.Get("/sports/new", adminSportsHandler.FormView)
			r.Post("/sports/new", adminSportsHandler.Create)
			
			// Users
			r.Get("/users", adminUsersHandler.View)
			r.Get("/users/edit/{id}", adminUsersHandler.EditFormView)
			r.Post("/users/edit/{id}", adminUsersHandler.Update)
			r.Post("/users/delete/{id}", adminUsersHandler.Delete)
			
			// Featured
			r.Get("/featured", adminFeaturedHandler.View)
			r.Get("/featured/new", adminFeaturedHandler.FormView)
			r.Post("/featured/new", adminFeaturedHandler.Create)
			r.Post("/featured/delete/{id}", adminFeaturedHandler.Delete)

			// Music
			r.Get("/music", adminMusicHandler.View)
			r.Get("/music/new", adminMusicHandler.FormView)
			r.Post("/music/new", adminMusicHandler.Create)
			r.Post("/music/delete/{id}", adminMusicHandler.Delete)

			// YouTube API Helper
			r.Get("/youtube/live-url", adminLiveTVHandler.FetchYouTubeLive)
			r.Post("/youtube/fetch-all", adminLiveTVHandler.FetchAllYouTubeLive)

			// Settings
			r.Get("/settings", adminSettingsHandler.View)
			r.Post("/settings", adminSettingsHandler.Update)

			// Categories
			r.Get("/categories", adminCategoriesHandler.View)
			r.Post("/categories/new", adminCategoriesHandler.Create)
			r.Post("/categories/delete", adminCategoriesHandler.Delete)
		})
	})
    
	r.Get("/ws/party", websockets.ServeWS(hub))

	port := os.Getenv("PORT")
	if port == "" {
		port = "9876"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("🚀 SheedBox Backend initialized and running", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServe failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for terminate signal
	<-stop

	slog.Info("Shutting down server gracefully...")

	// Timeout context for shutdown draining (30 seconds)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	} else {
		slog.Info("Server exited cleanly!")
	}
}

// registerAPIRoutes mounts all public and private API endpoints onto the router.
func registerAPIRoutes(
	r chi.Router,
	authHandler *handlers.AuthHandler,
	gamificationHandler *handlers.GamificationHandler,
	profileHandler *handlers.ProfileHandler,
	watchlistHandler *handlers.WatchlistHandler,
	settingsHandler *handlers.SettingsHandler,
	upNextHandler *handlers.UpNextHandler,
	featuredHandler *handlers.FeaturedHandler,
	movieHandler *handlers.MovieHandler,
	seriesHandler *handlers.SeriesHandler,
	livetvHandler *handlers.LiveTVHandler,
	sportsHandler *handlers.SportsHandler,
	musicHandler *handlers.MusicHandler,
	recommendationHandler *handlers.RecommendationHandler,
) {
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)
	r.Post("/auth/google", authHandler.GoogleLogin)

	// Settings and Up Next
	r.Get("/settings/menu", settingsHandler.GetSettings)
	r.Get("/content/up-next", upNextHandler.GetUpNext)

	// Public catalog routes
	r.Get("/featured", featuredHandler.GetFeatured)
	r.Get("/movies", movieHandler.GetMovies)
	r.Get("/movies/{id}", movieHandler.GetMovieDetail)
	r.Get("/series", seriesHandler.GetSeries)
	r.Get("/series/{id}", seriesHandler.GetSeriesDetail)
	r.Get("/live-tv", livetvHandler.GetLiveChannels)
	r.Get("/sports", sportsHandler.GetSportsEvents)
	r.Get("/music", musicHandler.GetMusic)
	r.Get("/music/{id}", musicHandler.GetMusicDetail)

	// Protected endpoints requiring JWT validation
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTMiddleware)

		// Profiles
		r.Get("/profiles", profileHandler.GetUserProfiles)
		r.Post("/profiles", profileHandler.CreateProfile)
		r.Put("/profiles/{id}", profileHandler.UpdateProfile)
		r.Delete("/profiles/{id}", profileHandler.DeleteProfile)

		r.Get("/recommendations", recommendationHandler.GetRecommendations)

		r.Post("/movies/resume", movieHandler.ResumePlayback)
		r.Get("/watchlists", watchlistHandler.GetUserWatchlists)
		r.Post("/gamification/trivia-reward", gamificationHandler.TriviaReward)
	})
}
