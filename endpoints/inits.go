package endpoints

import (
	"os"
	"time"

	"mongomini/agra/moncore"
)

var (
	Inited  bool = false
	initing bool = false

	Moncore *moncore.Moncore
)

// Initialize All
func InitAll() {

	for initing {
		time.Sleep(100 * time.Millisecond)
	}

	if Inited {
		return
	}

	initing = true

	_InitEnvArgs()

	_InitMongoDB()

	_InitEndpoints()

	Inited = true
	initing = false

	// Print("Collection names : " + strings.Join(Moncore.Database("users").ListCollection(), ", "))

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

	MC, err := moncore.InitMongo(Mongo_proto + Mongo_user + ":" + Mongo_pass + "@" + Mongo_host)

	if err != nil {
		panic(err)
	}

	Moncore = MC

}

// Initialize the endpoints
func _InitEndpoints() {

	API_Endpoints = append(API_Endpoints,
		API_Call_Handler_Prefix(`api/auth/`, API_Auth),
		API_Call_Handler_Prefix(`api/hello/`, API_Hello),
		API_Call_Handler_Exact(`mini/hello/([^/]+)/([^/]+)/`, API_Hello),
		API_Call_Handler_Prefix(`mini/hello/(.*)`, API_Hello),
		API_Call_Handler_Exact(`mini/ls/([^/]+)/`, API_List_Collections),
	)
}
