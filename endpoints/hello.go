package endpoints

import "strings"

// Test for API calls
func API_Hello(C *APICall) {
	C.WriteString("<h1> API_Hello! </h1> \n <br><br> \n Path : " +
		C.Path + " \n <br> \n Path Params : " + strings.Join(C.Params, ", ") +
		" \n <br> \n Method : " + C.Method() + " \n <br> \n Body : \n <br> \n " +
		C.BodyToString() + " \n <br> ")

	if len(C.Params) == 2 {
		C.WriteString("<br> \n <h2> Let's go, Captain " + C.Params[0] + " " + C.Params[1] + "! </h2> \n <br>")
	}
}

func API_List_Collections(C *APICall) {

	if len(C.Params) == 1 {
		C.WriteString("Collection names : " + strings.Join(Moncore.Database(C.Params[0]).ListCollection(), ", "))
	}
}
