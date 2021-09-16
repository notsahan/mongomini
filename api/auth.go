package handler

import (
	"net/http"
	"strconv"
)

func RawAuthHandler(w http.ResponseWriter, r *http.Request) {
	ServeRequest(w, r)
}

func API_POST_Auth(path string, body []byte) Response {
	return StringResponse("API POST recieved for api/auth/" + path + " with " + strconv.Itoa(len(body)) + " bytes of body")
}

func API_GET_Auth(path string) Response {
	return StringResponse("API GET recieved. Path : api/auth/" + path)
}
