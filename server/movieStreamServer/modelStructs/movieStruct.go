package modelStructs

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Movie struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ImdbID     string             `bson:"imdb_id" json:"imdb_id" validate:"required"`
	Title      string             `bson:"title" json:"title" validate:"required,min=2,max=500"`
	PosterPath string             `bson:"poster_path" json:"poster_path" validate:"required,url"`
	YouTubeID  string             `bson:"youtube_id" json:"youtube_id" validate:"required"`
	Genre      []Genre            `bson:"genre" json:"genre" validate:"required,dive"`
	AdminScore string             `bson:"admin_score" json:"admin_score"`
	Ranking    Ranking            `bson:"ranking" json:"ranking" validate:"required"`
}

type Genre struct {
	GenreID   int    `bson:"genre_id" json:"genre_id" validate:"required"`
	GenreName string `bson:"genre_name" json:"genre_name" validate:"required,min=2,max=100"`
}

type Ranking struct {
	RankingValue int    `bson:"ranking_value" json:"ranking_value" validate:"required"`
	RankingName  string `bson:"ranking_name" json:"ranking_name" validate:"required"`
}
