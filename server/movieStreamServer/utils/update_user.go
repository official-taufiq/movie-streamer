package utils

import (
	"context"
	"time"

	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/database"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func AddToken(userId, jwt, refresh, dbName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := database.OpenCollection("users", dbName)
	updateData := bson.M{
		"$set": bson.M{
			"token":         jwt,
			"refresh_token": refresh,
			"updated_at":    time.Now().UTC(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"user_id": userId}, updateData)
	if err != nil {
		return err
	}
	return nil

}
