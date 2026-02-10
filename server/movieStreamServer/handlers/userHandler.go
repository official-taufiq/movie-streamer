package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/firebase/genkit/go/genkit"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/database"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/modelStructs"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Config struct {
	JwtSecret  string
	DbName     string
	BasePrompt string
	ApiKey     string
	Genkit     *genkit.Genkit
	MovieLimit int64
}

func (cfg Config) AddUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var user modelStructs.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, fmt.Sprintf("error decoding user data:%v", err), http.StatusBadRequest)
		return
	}

	if err := validate.Struct(user); err != nil {
		http.Error(w, fmt.Sprintf("Error validation failed: %v", err), http.StatusBadRequest)
		return
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("error hashing password: %v", err), http.StatusInternalServerError)
		return
	}
	collection := database.OpenCollection("users", cfg.DbName)

	count, err := collection.CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to check existing user: %v", err), http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "user with this email already exists", http.StatusConflict)
		return
	}

	user.Password = hashedPassword
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.UserID = bson.NewObjectID().Hex()

	res, err := collection.InsertOne(ctx, user)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error adding user: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (cfg Config) LoginUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var userLogin modelStructs.UserLogin

	if err := json.NewDecoder(r.Body).Decode(&userLogin); err != nil {
		http.Error(w, fmt.Sprintf("error decoding body: %v", err), http.StatusInternalServerError)
		return
	}

	var user modelStructs.User

	collection := database.OpenCollection("users", cfg.DbName)

	if err := collection.FindOne(ctx, bson.M{"email": userLogin.Email}).Decode(&user); err != nil {
		http.Error(w, "Invalid Email", http.StatusUnauthorized)
		return
	}

	if err := utils.CheckPasswordAndHash(userLogin.Password, user.Password); err != nil {
		http.Error(w, "Wrong Password", http.StatusUnauthorized)
		return
	}

	hashedPass, err := utils.MakeJwt(user.UserID, cfg.JwtSecret, user.Role, time.Hour)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating JWT: %v", err), http.StatusInternalServerError)
		return
	}

	refreshToken, err := utils.MakeRefreshToken()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating Refresh Token: %v", err), http.StatusInternalServerError)
		return
	}

	if err := utils.AddToken(user.UserID, hashedPass, refreshToken, cfg.DbName); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update token: %v", err), http.StatusInternalServerError)
		return
	}

	userRes := modelStructs.UserResponse{
		UserID:         user.UserID,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Email:          user.Email,
		Role:           user.Role,
		Token:          hashedPass,
		RefreshToken:   refreshToken,
		FavoriteGenres: user.FavoriteGenres,
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userRes)

}
