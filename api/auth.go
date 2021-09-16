package handler

import (
	"mongomini/endpoints"
	"net/http"
)

func RawAuthHandler(w http.ResponseWriter, r *http.Request) {
	endpoints.ServeRequest(w, r)
}
