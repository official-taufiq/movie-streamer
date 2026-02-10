package utils

import (
	"context"
	"errors"
	"time"

	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func GetUserFavGenre(userId, dbName string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := database.OpenCollection("users", dbName)

	opts := options.FindOne().SetProjection(bson.M{
		"favorite_genres.genre_name": 1,
		"_id":                        0,
	})

	var result bson.M
	err := collection.FindOne(ctx, bson.M{"user_id": userId}, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []string{}, nil
		}
		return nil, err
	}

	favGenre, ok := result["favorite_genres"].(primitive.A)
	if !ok {
		return []string{}, errors.New("Unable to retrieve favorite genres")
	}

	var genres []string

	for _, item := range favGenre {
		if genreMap, ok := item.(bson.M); ok {
			if name, ok := genreMap["genre_name"].(string); ok {
				genres = append(genres, name)
			}
		}
	}

	return genres, nil
}
