package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func DBinstance(uri string) error {

	conn := options.Client().ApplyURI(uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error

	client, err = mongo.Connect(ctx, conn)
	if err != nil {
		return err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	return nil

}

func OpenCollection(collectionName, dbName string) *mongo.Collection {
	return client.Database(dbName).Collection(collectionName)
}
