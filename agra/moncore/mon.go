package moncore

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DefaultContextTimout time.Duration = 10 * time.Second
)

type Moncore struct {
	client *mongo.Client
}

func DefaultContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), DefaultContextTimout)
	return ctx
}

// MonCore is a wrapper around mongo.Client.
// URL is the connection string like 'Mongo_proto + Mongo_user + ":" + Mongo_pass + "@" + Mongo_host'
func InitMongo(url string) (*Moncore, error) {

	clientOptions := options.Client().ApplyURI(url)

	client, err := mongo.Connect(DefaultContext(), clientOptions)
	if err != nil {
		return nil, err
	}

	ping_err := client.Ping(DefaultContext(), nil)
	if ping_err != nil {
		return nil, err
	}

	Print("MongoDB client connected")
	return &Moncore{client: client}, nil
}

func (MC *Moncore) Disconnect() error {
	return MC.client.Disconnect(DefaultContext())
}

func (MC *Moncore) Database(name string) *MonDatabase {
	db := MC.client.Database(name)
	return &MonDatabase{db: db}
}

type MonDatabase struct {
	db *mongo.Database
}

func (MD *MonDatabase) Collection(name string) *MonCollection {
	return &MonCollection{col: MD.db.Collection(name)}
}

func (MD *MonDatabase) ListCollection() []string {
	names, err := MD.db.ListCollectionNames(DefaultContext(), bson.M{})

	if CheckError(err) {
		return nil
	}
	return names
}

type MonCollection struct {
	col *mongo.Collection
}
