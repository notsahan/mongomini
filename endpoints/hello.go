package endpoints

import "strings"

// Test for API calls
func API_Hello(C *APICall) {
	C.WriteString("<h1> API_Hello! </h1> <br><br> Path : " +
		C.Path + " <br> Path Params : " + strings.Join(C.Params, ", ") +
		" <br> Method : " + C.Method() + " <br> Body : <br> " +
		C.BodyToString() + " <br> ")

	if len(C.Params) == 2 {
		C.WriteString("<br> <h2> Let's go, Captain " + C.Params[0] + " " + C.Params[1] + "! </h2> <br>")
	}
}
