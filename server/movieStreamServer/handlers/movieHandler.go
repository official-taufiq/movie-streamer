package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/database"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/modelStructs"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/utils"
	"github.com/tmc/langchaingo/llms/openai"
	"go.mongodb.org/mongo-driver/v2/bson"
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
	imdbId := r.PathValue("imdb_id")

	req := struct {
		AdminReview string `json:"admin_review"`
	}{}

	res := struct {
		AdminReview string `json:"admin_review"`
		Ranking     string `json:"ranking"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decoding request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	llmRes, rankingValue, err := GetReviewRanking(cfg.ApiKey, cfg.BasePrompt, cfg.DbName, req.AdminReview)
	if err != nil {
		http.Error(w, "")
	}

}

func GetReviewRanking(apiKey, basePrompt, dbName, admin_review string) (string, int, error) {
	rankings, err := utils.GetRankings(dbName)
	if err != nil {
		return "", 0, err
	}

	str := ""

	for _, ranking := range rankings {
		if ranking.RankingValue != 999 {
			str = str + ranking.RankingName + ","
		}
	}

	str = strings.Trim(str, ",")

	llm, err := openai.New(openai.WithToken(apiKey))
	if err != nil {
		return "", 0, err
	}

	prompt := strings.Replace(basePrompt, "{rankings}", str, 1)

	res, err := llm.Call(context.Background(), prompt+admin_review)
	if err != nil {
		return "", 0, err
	}

	rankingValue := 0

	for _, ranking := range rankings {
		if ranking.RankingName == res {
			rankingValue = ranking.RankingValue
			break
		}
	}
	return res, rankingValue, nil
}
