package utils

import (
	"context"
	"time"

	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/database"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/modelStructs"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func GetRankings(dbName string) ([]modelStructs.Ranking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var rankings []modelStructs.Ranking
	collection := database.OpenCollection("rankings", dbName)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &rankings); err != nil {
		return nil, err
	}
	return rankings, nil
}
