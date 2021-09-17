package endpoints

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Inited      bool = false
	initing     bool = false
	MongoClient *mongo.Client
)

func InitAll() {

	for initing {
		time.Sleep(100 * time.Millisecond)
	}

	if Inited {
		return
	}

	initing = true

	_InitEnvArgs()

	// _InitMongoDB()

	_InitEndpoints()

	Inited = true
	initing = false
}

// Initialize from the environment variables
func _InitEnvArgs() {
	if envarg := os.Getenv("Mongo_proto"); len(envarg) != 0 {
		Mongo_proto = envarg
	}

	if envarg := os.Getenv("Mongo_user"); len(envarg) != 0 {
		Mongo_user = envarg
	}

	if envarg := os.Getenv("Mongo_pass"); len(envarg) != 0 {
		Mongo_pass = envarg
	}

	if envarg := os.Getenv("Mongo_host"); len(envarg) != 0 {
		Mongo_host = envarg
	}

	if envarg := os.Getenv("Mongo_DBName"); len(envarg) != 0 {
		Mongo_DBName = envarg
	}
}

// Initialize the mongo client
func _InitMongoDB() {
	// "minivercel:<password>@minicluster.hybfy.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"

	clientOptions := options.Client().
		ApplyURI(Mongo_proto + Mongo_user + ":" + Mongo_pass + "@" + Mongo_host)

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	MongoClient = client

}

func _InitEndpoints() {

	API_Endpoints = append(API_Endpoints,
		API_Call_Handler_Prefix(`api/auth/`, API_Auth),
		API_Call_Handler_Prefix(`api/hello/`, API_Hello),
		API_Call_Handler_Exact(`mini/hello/([^/]+)/([^/]+)/`, API_Hello),
		API_Call_Handler_Prefix(`mini/hello/(.*)`, API_Hello),
	)
}
