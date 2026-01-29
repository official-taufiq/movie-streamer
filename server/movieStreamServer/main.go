package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/database"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/handlers"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Unable to find .env file")
	}

	if err = database.DBinstance(); err != nil {
		log.Fatalf("Mongo connection failed: %v", err)
	}
	mux := http.NewServeMux()
	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.HandleFunc("GET /movies", handlers.GetMovieHandler)

	fmt.Println("Starting movie stream server on :8080")
	srv.ListenAndServe()
}
