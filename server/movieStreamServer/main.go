package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/joho/godotenv"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/database"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/handlers"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/middlewares"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Unable to find .env file")
	}
	secret := os.Getenv("JWT_SECRET")
	uri := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("DATABASE_NAME")
	basePrompt := os.Getenv("BASE_PROMPT")
	apiKeyGroq := os.Getenv("API_KEY_GROQ")
	apiKeyGemini := os.Getenv("API_KEY_GEMINI")
	movieLimit, err := strconv.ParseInt(os.Getenv("RECOMMENDED_MOVIE_LIMIT"), 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	g := genkit.Init(context.Background(), genkit.WithPlugins(&googlegenai.GoogleAI{APIKey: apiKeyGemini}),
		genkit.WithDefaultModel("googleai/gemini-2.5-flash"))

	authCfg := middlewares.Config{
		JwtSecret: secret,
	}
	handlerCfg := handlers.Config{
		JwtSecret:  secret,
		DbName:     dbName,
		BasePrompt: basePrompt,
		ApiKey:     apiKeyGroq,
		Genkit:     g,
		MovieLimit: movieLimit,
	}

	if err = database.DBinstance(uri); err != nil {
		log.Fatalf("Mongo connection failed: %v", err)
	}
	mux := http.NewServeMux()
	srv := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	mux.Handle("GET /movie/{imdb_id}", authCfg.AuthMiddleware(http.HandlerFunc(handlerCfg.GetOneMovieHandler)))
	mux.Handle("POST /addmovie", authCfg.AuthMiddleware(http.HandlerFunc(handlerCfg.AddMovie)))
	mux.Handle("GET /recmovies", authCfg.AuthMiddleware(http.HandlerFunc(handlerCfg.GetRecommendations)))
	mux.HandleFunc("GET /movies", handlerCfg.GetMovieHandler)
	mux.HandleFunc("POST /register", handlerCfg.AddUser)
	mux.HandleFunc("POST /login", handlerCfg.LoginUser)
	mux.HandleFunc("POST /adminreview/{imdb_id}", handlerCfg.AdminReview)

	fmt.Println("Starting movie stream server on :8080")
	log.Fatal(srv.ListenAndServe())
}
