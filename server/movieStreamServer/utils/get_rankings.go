package utils

import (
	"context"
	"strings"
	"time"

	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/database"
	"github.com/official-taufiq/movie-streamer/server/movieStreamServer/modelStructs"
	"github.com/tmc/langchaingo/llms/openai"
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

func GetReviewRanking(apiKey, basePrompt, dbName, admin_review string) (string, int, error) {
	rankings, err := GetRankings(dbName)
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
