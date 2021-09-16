package handler

import (
	"net/http"
)

var (
	Mongo_proto  string = "mongodb+srv://"
	Mongo_user   string = "username"
	Mongo_pass   string = "apsppasss"
	Mongo_host   string = "minicluster.hybfy.mongodb.net"
	Mongo_DBName string = "database"
)

func CredsHandler(w http.ResponseWriter, r *http.Request) {
	ServeRequest(w, r)
}
