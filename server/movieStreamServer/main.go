package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/database"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/handlers"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/middlewares"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Unable to find .env file")
	}
	secret := os.Getenv("JWT_SECRET")
	uri := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("DATABASE_NAME")

	authCfg := middlewares.Config{
		JwtSecret: secret,
	}
	handlerCfg := handlers.Config{
		JwtSecret: secret,
		DbName:    dbName,
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
	mux.HandleFunc("GET /movies", handlerCfg.GetMovieHandler)
	mux.HandleFunc("POST /register", handlerCfg.AddUser)
	mux.HandleFunc("POST /login", handlerCfg.LoginUser)

	fmt.Println("Starting movie stream server on :8080")
	log.Fatal(srv.ListenAndServe())
}
