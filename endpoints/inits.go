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

type API_POST_Recieved func(string, []byte) Response
type API_GET_Recieved func(string) Response

type Response struct {
	Body    []byte
	Headers map[string]string
}

func BodyResponse(body []byte) Response {
	return Response{body, map[string]string{}}
}

func StringResponse(body string) Response {
	return Response{[]byte(body), map[string]string{}}
}

func InitAll() {

	for initing {
		time.Sleep(100 * time.Millisecond)
	}

	if Inited {
		return
	}

	initing = true

	InitEnvArgs()

	InitMongoDB()

	Inited = true
	initing = false
}

// Initialize from the environment variables
func InitEnvArgs() {
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
func InitMongoDB() {
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
