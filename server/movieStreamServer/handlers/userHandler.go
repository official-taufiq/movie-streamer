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
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func AddUser(w http.ResponseWriter, r *http.Request) {
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

	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("error hashing password: %v", err), http.StatusInternalServerError)
		return
	}
	collection := database.OpenCollection("users")

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

func LoginUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var userLogin modelStructs.UserLogin

	if err := json.NewDecoder(r.Body).Decode(&userLogin); err != nil {
		http.Error(w, fmt.Sprintf("error decoding body: %v", err), http.StatusInternalServerError)
		return
	}

	var user modelStructs.User

	collection := database.OpenCollection("users")

	if err := collection.FindOne(ctx, bson.M{"email": userLogin.Email}).Decode(&user); err != nil {
		http.Error(w, "Invalid Email", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userLogin.Password)); err != nil {
		http.Error(w, "Wrong Password", http.StatusUnauthorized)
		return
	}

}
