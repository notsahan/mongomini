package handler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var API_POSTs map[string]API_POST_Recieved = map[string]API_POST_Recieved{"api/auth": api_POST_Auth}

var API_GETs map[string]API_GET_Recieved = map[string]API_GET_Recieved{"api/auth": api_GET_Auth}

func ServeRequest(w http.ResponseWriter, r *http.Request) {

	urlpath := strings.TrimPrefix(r.URL.Path, "/")

	switch r.Method {
	case "GET":

		for apiPath, rapi := range API_GETs {
			if strings.HasPrefix(urlpath, apiPath) {
				resp := rapi(strings.TrimPrefix(urlpath, apiPath+"/"))
				if resp.Headers != nil {
					for k, v := range resp.Headers {
						w.Header().Set(k, v)
					}
				}
				if len(resp.Body) > 0 {
					w.Write(resp.Body)
				}
				return
			}
		}

		w.Write([]byte("Path " + urlpath + " not found"))

		// // If no API
		// fmt.Println("Serving " + urlpath)

		// http.ServeFile(w, r, path.Join(rootDir, urlpath))

	case "POST":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		println("POST " + urlpath + " body length " + strconv.Itoa(len(body)))

		for apiPath, rapi := range API_POSTs {
			if strings.HasPrefix(urlpath, apiPath) {
				resp := rapi(strings.TrimPrefix(urlpath, apiPath+"/"), body)

				if resp.Headers != nil {
					for k, v := range resp.Headers {
						w.Header().Set(k, v)
					}
				}
				if len(resp.Body) > 0 {
					w.Write(resp.Body)
				}
				return
			}
		}

		w.Write([]byte("Path " + urlpath + " not found"))

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}
