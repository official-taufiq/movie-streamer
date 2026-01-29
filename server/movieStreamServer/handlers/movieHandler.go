package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/database"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/modelStructs"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func GetMovieHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	w.Header().Set("Content-Type", "application/json")

	var movies []modelStructs.Movie

	collection := database.OpenCollection("movies")

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error Find:%s", err), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &movies); err != nil {
		http.Error(w, fmt.Sprintf("Error Cursor:%s", err), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(movies); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
