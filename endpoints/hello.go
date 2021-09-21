package endpoints

import (
	"errors"
	"mongomini/agra/moncore"
	"strconv"
	"strings"
	"time"
)

// Test for API calls
func API_Hello(C *APICall) {
	C.HTMLBegin()

	C.WriteString("<h1> API_Hello! </h1> \n <br><br> \n Path : " +
		C.Path + " \n <br> \n Path Params : " + strings.Join(C.Params, ", ") +
		" \n <br> \n Method : " + C.Method() + " \n <br> \n Body : \n <br> \n " +
		C.BodyToString() + " \n <br> ")

	if len(C.Params) == 2 {
		C.WriteString("<br> \n <h2> Let's go, Captain " + C.Params[0] + " " + C.Params[1] + "! </h2> \n <br>")
	}

	C.HTMLEnd()

}

func API_List_Collections(C *APICall) {

	C.HTMLBegin()

	if len(C.Params) == 1 {
		C.WriteString("Collection names : " + strings.Join(Moncore.Database(C.Params[0]).ListCollectionNames(), ", "))

		var cols = Moncore.Database(C.Params[0]).ListCollections()

		C.WriteString(" <br> <br> Collection count : " + strconv.Itoa(len(cols)) + " <br> ")

	} else {
		C.WriteError("Bad Request", errors.New("endpoints : mini/ls/<db>"), 400)
		return
	}

	C.HTMLEnd()

}

func API_List_Documents(C *APICall) {

	if len(C.Params) == 2 {
		Docs := Moncore.Database(C.Params[0]).Collection(C.Params[1]).Query(moncore.Filter_MatchAll{})
		C.WriteJSON(Docs)

	} else {
		C.WriteError("Bad Request", errors.New("endpoint : mini/ls/<db>/<collection>"), 400)
		return
	}

}

func API_Set_Document(C *APICall) {

	C.HTMLBegin()

	var doc interface{}

	if len(C.Params) == 3 {

		if C.Method() == "POST" {
			doc = C.BodyToString()
		} else {
			mapdoc := map[string]string{"Created": time.Now().UTC().String()}
			doc = mapdoc
		}

	} else if len(C.Params) == 5 {
		key := C.Params[3]
		value := C.Params[4]

		mapdoc := map[string]string{key: value}
		doc = mapdoc
	} else {
		C.WriteError("Bad Request", errors.New("must have only one parameter. Endpoints : mini/set/<db</<collection>/<dockey> or mini/set/<db</<collection>/<dockey>/<key>/<value>"), 400)
		return
	}

	Col := Moncore.Database(C.Params[0]).Collection(C.Params[1])

	NewDocKey := Col.Set(C.Params[2], doc)

	C.WriteString("<pre>")

	if len(NewDocKey) == 0 {
		C.WriteString("Updated document")
	} else {
		C.WriteString("Inserted new : " + NewDocKey)

	}

	C.WriteString("</pre>")

	C.HTMLEnd()

}
