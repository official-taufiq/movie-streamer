package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func DBinstance(uri string) error {

	conn := options.Client().ApplyURI(uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error

	Client, err = mongo.Connect(ctx, conn)
	if err != nil {
		return err
	}
	if err := Client.Ping(ctx, nil); err != nil {
		return err
	}

	return nil

}

func OpenCollection(collectionName, dbName string) *mongo.Collection {
	return Client.Database(dbName).Collection(collectionName)
}
