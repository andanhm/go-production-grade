package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Config interface {
	URIBuilder() (string, string)
}

func New(config Config) (*mongo.Database, error) {
	uri, dbName := config.URIBuilder()
	client, err := mongo.Connect(
		context.Background(),
		options.Client().SetConnectTimeout(time.Second*10),
		options.
			Client().
			ApplyURI(uri).
			SetLoggerOptions(options.
				Logger().
				SetComponentLevel(
					options.LogComponentConnection,
					options.LogLevelInfo,
				).
				SetSink(&Logger{}),
			),
	)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return client.Database(dbName), nil
}
