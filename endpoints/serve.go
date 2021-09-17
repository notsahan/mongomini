package endpoints

// Self-contained handler for API calls.

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

// You can use API_Call_Handler_*** templates to easily define API endpoints.
// First regex match in the order of the array will be used, and no other matches will be attempted.
//
// URL paths are standardised to start without a slash and end with a slash before matching.
// For example, "https://api.example.com/api/user/123" will become "api/user/123/" before matching.
//
// This is used by the ServeRequest function.
// You can also add API Call Handlers dynamically by appending this array.
var API_Endpoints []API_Call_Handler = []API_Call_Handler{}

// Handles if there are no matching endpoints found in API_Endpoints
var API_Not_Found_Handler func(*APICall)

// ServeRequest is the main entry point for the API.
func ServeRequest(w http.ResponseWriter, r *http.Request) {

	// Standarize the path
	urlpath := strings.TrimPrefix(r.URL.Path, "/")
	if !strings.HasSuffix(urlpath, "/") {
		urlpath += "/"
	}

	for _, handler := range API_Endpoints {
		rmatches := handler.Path.FindAllStringSubmatch(urlpath, 1)
		if len(rmatches) == 0 || len(rmatches[0]) == 0 || handler.Handler == nil {
			continue
		}

		// We have a match, so call the handler
		pathParams := rmatches[0][1:]
		C := &APICall{HTTPWriter: &w, HTTPRequest: r, Path: urlpath, Params: pathParams}
		handler.Handler(C)

		return
	}

	// If we get here, we didn't find a handler
	if API_Not_Found_Handler != nil {
		C := &APICall{HTTPWriter: &w, HTTPRequest: r, Path: urlpath, Params: nil}
		API_Not_Found_Handler(C)
	} else {
		http.NotFound(w, r)
	}

}

type APICall struct {
	HTTPWriter  *http.ResponseWriter
	HTTPRequest *http.Request

	// Request path without host or protocole. Begins without a slash, ends with a slash
	Path string

	// Parameters in the path. Created from submatches of regexp from hanlder (see API_Call_Handler).
	// Will be empty if no parameters were used.
	Params []string
}

type API_Call_Handler struct {

	// Handler is the function that will be called when the path matches.
	Handler func(*APICall)

	// Path is the regex that will be used to match the path.
	// Captured regex groups (Submatches) will be passed to the APICall handler as Params.
	// Path must not begin with a slash.
	Path *regexp.Regexp
}

// Method specifies the HTTP method (GET, POST, PUT, etc.).
// For client requests, an empty string means GET.
func (c *APICall) Method() string {
	return c.HTTPRequest.Method
}

// Request body
func (c *APICall) Body() []byte {

	body, err := ioutil.ReadAll(c.HTTPRequest.Body)
	if err != nil {
		PrintErrorMsg("Error reading body: ", err)
		http.Error(*c.HTTPWriter, "can't read body", http.StatusBadRequest)
		return nil
	}

	return body
}

// Request body as string
func (c *APICall) BodyToString() string {
	return string(c.Body())
}

// Deserializing request body to struct type
func (c *APICall) BodyJsonToStruct(Type interface{}) {
	if err := json.Unmarshal(c.Body(), Type); err != nil {
		PrintErrorMsg("Error unmarshaling body: ", err)
		c.WriteError("Error deserializing body", err, http.StatusUnprocessableEntity)
	}
}

// Write binary to response
func (c *APICall) Write(Bin []byte) {
	(*c.HTTPWriter).Write(Bin)
}

// Write string to response
func (c *APICall) WriteString(S string) {
	c.Write([]byte(S))
}

// Write Object to response as JSON
func (c *APICall) WriteJSON(Obj interface{}) {
	J, JErr := json.Marshal(Obj)
	if JErr != nil {
		PrintErrorMsg("Error marshalling JSON: ", JErr)
		c.WriteError("Error Writing to JSON ", JErr, http.StatusUnprocessableEntity)
		return
	}

	c.Write(J)
}

// Write http.Status___ code to response
func (c *APICall) WriteStatus(HttpStatusCode int) {
	(*c.HTTPWriter).WriteHeader(HttpStatusCode)
}

// Get Header from requset
func (c *APICall) GetHeader(key string) string {
	return c.HTTPRequest.Header.Get(key)
}

// Set header to response
func (c *APICall) SetHeader(key string, value string) {
	(*c.HTTPWriter).Header().Set(key, value)
}

// Write error to response
func (c *APICall) WriteError(PrefixMsg string, Err error, HttpStatusCode int) {
	PrintErrorMsg("WriteError: ", Err)
	http.Error(*c.HTTPWriter, PrefixMsg+Err.Error(), HttpStatusCode)
}

// Get cookie from request
func (c *APICall) GetCookie(key string) string {
	cookie, err := c.HTTPRequest.Cookie(key)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// Set cookie to response
func (c *APICall) SetCookie(key string, value string) {
	http.SetCookie(*c.HTTPWriter, &http.Cookie{
		Name:  key,
		Value: value,
	})
}

// Creates case insensitive Regex matcher begining with 'prefix'. Prefix is also considered as a regex.
// Path must not begin with a slash ; Path must end with a slash.
// For example, "api/auth/" is a valid path prefix.
func API_Call_Handler_Prefix(prefix string, Handler func(*APICall)) API_Call_Handler {
	r := regexp.MustCompile(`(?i)^` + prefix)
	return API_Call_Handler{Handler, r}
}

// Creates case insensitive Regex matcher containing only 'path'. Path is also considered as a regex.
// Path must not begin with a slash. Path must end with a slash.
// For example, "api/auth/" is a valid path.
func API_Call_Handler_Exact(path string, Handler func(*APICall)) API_Call_Handler {
	r := regexp.MustCompile(`(?i)^` + path + `$`)
	return API_Call_Handler{Handler, r}
}
