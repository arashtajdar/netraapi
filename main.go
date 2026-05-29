package main

import (
	"log"
	"net/http"
	"os"

	"netra-api/config"
	"netra-api/handlers"
	"netra-api/middleware"
	"netra-api/websockets"

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

		// Protected endpoints requiring JWT validation
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTMiddleware)
			r.Get("/movies", handlers.GetMovies)
			r.Post("/movies/resume", handlers.ResumePlayback)
			r.Post("/gamification/trivia-reward", handlers.TriviaReward)
		})
	})
    
	r.Get("/ws/party", websockets.ServeWS(hub))

	port := os.Getenv("PORT")
	if port == "" {
		port = "9876"
	}

	log.Printf("🚀 Netra Backend initialized and running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
