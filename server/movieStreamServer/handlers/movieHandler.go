package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/database"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/modelStructs"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"net/http"
	"time"
)

var validate = validator.New()

func (cfg Config) GetMovieHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	w.Header().Set("Content-Type", "application/json")

	var movies []modelStructs.Movie
	collection := database.OpenCollection("movies", cfg.DbName)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Collection not found:%s", err), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &movies); err != nil {
		http.Error(w, fmt.Sprintf("Error Cursor:%s", err), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(movies); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (cfg Config) GetOneMovieHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	w.Header().Set("Content-Type", "application/json")

	imdbID := r.PathValue("imdb_id")

	var movie modelStructs.Movie
	collection := database.OpenCollection("movies", cfg.DbName)

	if err := collection.FindOne(ctx, bson.M{"imdb_id": imdbID}).Decode(&movie); err != nil {
		http.Error(w, fmt.Sprintf("Movie not found:%v", err), http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(movie); err != nil {
		http.Error(w, fmt.Sprintf("error encoding response: %v", err), http.StatusInternalServerError)
	}
}

func (cfg Config) AddMovie(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var movie modelStructs.Movie

	err := json.NewDecoder(r.Body).Decode(&movie)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, fmt.Sprintf("error decoding body: %v", err), http.StatusInternalServerError)
		return
	}

	if err := validate.Struct(movie); err != nil {
		http.Error(w, fmt.Sprintf("error: Validation failed, details: %v", err), http.StatusBadRequest)
		return
	}

	collection := database.OpenCollection("movies", cfg.DbName)
	res, err := collection.InsertOne(ctx, movie)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error adding movie: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (cfg Config) AdminReview(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	imdbId := r.PathValue("imdb_id")

	req := struct {
		AdminReview string `json:"admin_review"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decoding request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	llmRes, rankingValue, err := utils.GetReviewRanking(cfg.ApiKey, cfg.BasePrompt, cfg.DbName, req.AdminReview)
	if err != nil {
		http.Error(w, "Error getting a review ranking", http.StatusInternalServerError)
		return
	}

	collection := database.OpenCollection("movies", cfg.DbName)

	var movie modelStructs.Movie

	if err := collection.FindOne(ctx, bson.M{"imdb_id": imdbId}).Decode(&movie); err != nil {
		http.Error(w, fmt.Sprintf("Movie not found:%v", err), http.StatusNotFound)
		return
	}

	updateData := bson.M{
		"$set": bson.M{
			"admin_review": req.AdminReview,
			"ranking": bson.M{
				"ranking_value": rankingValue,
				"ranking_name":  llmRes,
			},
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"imdb_id": imdbId}, updateData)
	if err != nil {
		http.Error(w, "Error updating data", http.StatusInternalServerError)
		return
	}
	if result.MatchedCount == 0 {
		http.Error(w, "No movie found with the provided imdb ID", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		RankingName string `json:"ranking_name"`
		AdminReview string `json:"admin_review"`
	}{
		RankingName: req.AdminReview,
		AdminReview: llmRes,
	})
}
