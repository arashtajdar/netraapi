package main

import (
	"log"
	"net/http"
	"os"

	"netra-api/config"
	"netra-api/handlers"
	"netra-api/middleware"
	"netra-api/websockets"
	"netra-api/services/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found or error reading it")
	}

	config.ConnectDB()
	defer config.DB.Close()

	if os.Getenv("STORAGE_DRIVER") == "r2" {
		storage.ActiveProvider = storage.NewR2Storage(
			os.Getenv("R2_ACCOUNT_ID"),
			os.Getenv("R2_ACCESS_KEY_ID"),
			os.Getenv("R2_SECRET_ACCESS_KEY"),
			os.Getenv("R2_BUCKET_NAME"),
			os.Getenv("R2_PUBLIC_URL"),
		)
		log.Println("☁️ Cloudflare R2 Storage Provider Initialized")
	} else {
		storage.ActiveProvider = storage.NewLocalStorage(
			os.Getenv("STORAGE_LOCAL_DIR"),
			os.Getenv("STORAGE_LOCAL_URL"),
		)
		log.Println("📂 Local Storage Provider Initialized")
	}

	hub := websockets.NewHub()
	go hub.Run()

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/register", handlers.Register)
		r.Post("/auth/login", handlers.Login)
		r.Post("/auth/google", handlers.GoogleLogin)



		// Public catalog routes
		r.Get("/movies", handlers.GetMovies)
		r.Get("/movies/{id}", handlers.GetMovieDetail)
		r.Get("/series", handlers.GetSeries)
		r.Get("/series/{id}", handlers.GetSeriesDetail)
		r.Get("/live-tv", handlers.GetLiveChannels)
		r.Get("/sports", handlers.GetSportsEvents)

		// Protected endpoints requiring JWT validation
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTMiddleware)
			r.Post("/movies/resume", handlers.ResumePlayback)
			r.Get("/watchlists", handlers.GetUserWatchlists)
			r.Post("/gamification/trivia-reward", handlers.TriviaReward)
		})
	})

	// Admin UI Routes (Moved outside API to fix 404s)
	r.Route("/admin", func(r chi.Router) {
		r.Get("/", handlers.AdminDashboardView)
		r.Post("/upload", handlers.UploadMedia)
		r.Get("/movies", handlers.AdminMoviesView)
		r.Get("/movies/new", handlers.AdminMoviesFormView)
		r.Post("/movies/new", handlers.AdminMoviesCreate)
		r.Get("/series", handlers.AdminSeriesView)
		r.Get("/series/new", handlers.AdminSeriesFormView)
		r.Post("/series/new", handlers.AdminSeriesCreate)
		r.Get("/live-tv", handlers.AdminLiveTVView)
		r.Get("/live-tv/new", handlers.AdminLiveTVFormView)
		r.Post("/live-tv/new", handlers.AdminLiveTVCreate)
		r.Post("/live-tv/delete/{id}", handlers.AdminLiveTVDelete)
		r.Get("/sports", handlers.AdminSportsView)
		r.Get("/sports/new", handlers.AdminSportsFormView)
		r.Post("/sports/new", handlers.AdminSportsCreate)
	})
    
	r.Get("/ws/party", websockets.ServeWS(hub))

	port := os.Getenv("PORT")
	if port == "" {
		port = "9876"
	}

	log.Printf("🚀 Netra Backend initialized and running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
