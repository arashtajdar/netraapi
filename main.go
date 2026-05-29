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

	storageDriver := os.Getenv("STORAGE_DRIVER")
	if storageDriver == "r2" {
		storage.ActiveProvider = storage.NewR2Storage(
			os.Getenv("R2_ACCOUNT_ID"),
			os.Getenv("R2_ACCESS_KEY_ID"),
			os.Getenv("R2_SECRET_ACCESS_KEY"),
			os.Getenv("R2_BUCKET_NAME"),
			os.Getenv("R2_PUBLIC_URL"),
		)
	} else {
		localDir := os.Getenv("STORAGE_LOCAL_DIR")
		if localDir == "" { localDir = "./tmp/media" }
		localUrl := os.Getenv("STORAGE_LOCAL_URL")
		if localUrl == "" { localUrl = "http://127.0.0.1:9876/media" }
		storage.ActiveProvider = storage.NewLocalStorage(localDir, localUrl)
	}

	config.ConnectDB()
	defer config.DB.Close()

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

		// Admin UI Routes
		r.Route("/admin", func(r chi.Router) {
			r.Post("/upload", handlers.UploadMedia)
			r.Get("/", handlers.AdminDashboardView)
			r.Get("/movies", handlers.AdminMoviesView)
			r.Get("/movies/new", handlers.AdminMoviesFormView)
			r.Post("/movies/new", handlers.AdminMoviesCreate)
			r.Get("/series", handlers.AdminSeriesView)
			r.Get("/series/new", handlers.AdminSeriesFormView)
			r.Post("/series/new", handlers.AdminSeriesCreate)
			r.Get("/live-tv", handlers.AdminLiveTVView)
			r.Get("/live-tv/new", handlers.AdminLiveTVFormView)
			r.Post("/live-tv/new", handlers.AdminLiveTVCreate)
			r.Get("/sports", handlers.AdminSportsView)
			r.Get("/sports/new", handlers.AdminSportsFormView)
			r.Post("/sports/new", handlers.AdminSportsCreate)
		})

		// Protected endpoints requiring JWT validation
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTMiddleware)
			r.Get("/movies", handlers.GetMovies)
			r.Post("/movies/resume", handlers.ResumePlayback)
			
			// New Streaming Platform Routes
			r.Get("/series", handlers.GetSeries)
			r.Get("/live-tv", handlers.GetLiveChannels)
			r.Get("/sports", handlers.GetSportsEvents)
			r.Get("/watchlists", handlers.GetUserWatchlists)
			
			r.Post("/gamification/trivia-reward", handlers.TriviaReward)
		})
	})
    
	r.Get("/ws/party", websockets.ServeWS(hub))
	
	// Serve local media files if using local storage
	r.Handle("/media/*", http.StripPrefix("/media/", http.FileServer(http.Dir("./tmp/media"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "9876"
	}

	log.Printf("🚀 Netra Backend initialized and running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
