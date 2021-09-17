package handler

import (
	"mongomini/endpoints"
	"net/http"
)

func AllHandler(w http.ResponseWriter, r *http.Request) {
	endpoints.InitAll()
	endpoints.ServeRequest(w, r)
}
