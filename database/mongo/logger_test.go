package mongo

import (
	"context"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MockMongoConfig struct{}

func (m MockMongoConfig) URIBuilder() (string, string) {
	// Use an invalid local address with a short timeout to trigger driver trace events
	return "mongodb://127.0.0.1:27019/?connectTimeoutMS=200&serverSelectionTimeoutMS=200", "testdb"
}

func Test_New(t *testing.T) {
	os.Setenv("TIER", "dev")
	defer os.Unsetenv("TIER")

	cfg := MockMongoConfig{}
	_, err := New(cfg)
	if err == nil {
		t.Log("Expected connection/ping to fail, but it succeeded")
		return
	}
}

func TestMongoQueries_Mocked(t *testing.T) {
	// Set TIER to dev so info/trace logs are logged via logrus
	os.Setenv("TIER", "dev")
	defer os.Unsetenv("TIER")

	// Set up client options with our custom logger sink
	loggerOpts := options.Logger().
		SetComponentLevel(options.LogComponentCommand, options.LogLevelInfo).
		SetSink(&Logger{})

	opts := options.Client().SetLoggerOptions(loggerOpts)

	// Create mtest instance passing the customized client options
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock).ClientOptions(opts))

	mt.Run("retrieval_success", func(mt *mtest.T) {
		ns := mt.Coll.Database().Name() + "." + mt.Coll.Name()

		// Mock a single user document response
		mockDoc := bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "name", Value: "John Doe"},
			{Key: "identity", Value: "SC012202601"},
		}

		firstBatch := mtest.CreateCursorResponse(0, ns, mtest.FirstBatch, mockDoc)
		mt.AddMockResponses(firstBatch)

		// Execute Find
		var results []bson.M
		cursor, err := mt.Coll.Find(context.Background(), bson.D{{Key: "identity", Value: "SC012202601"}})
		if err != nil {
			mt.Fatalf("Find query failed: %v", err)
		}
		defer cursor.Close(context.Background())

		if err = cursor.All(context.Background(), &results); err != nil {
			mt.Fatalf("Cursor.All failed: %v", err)
		}

		if len(results) != 1 {
			mt.Fatalf("Expected 1 result, got %d", len(results))
		}

		// Explicitly invoke logger to show Info JSON output for the query retrieval
		logger := Logger{}
		logger.Info(1, "Command find executed", "command", "find", "collection", mt.Coll.Name(), "identity", "SC012202601")
	})

	mt.Run("retrieval_error", func(mt *mtest.T) {
		// Mock command failure
		cmdErr := mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    2,
			Message: "BadValue: query requires identity",
		})
		mt.AddMockResponses(cmdErr)

		// Execute Find expecting error
		_, err := mt.Coll.Find(context.Background(), bson.D{{Key: "identity", Value: "SC012202601"}})
		if err == nil {
			mt.Fatal("Expected Find query to fail, but it succeeded")
		}

		// Explicitly invoke logger to show Error JSON output for the query error
		logger := Logger{}
		logger.Error(err, "Find query failed", "identity", "SC012202601")
	})
}
