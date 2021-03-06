package endpoints

import (
	"errors"
	"mongomini/agra/moncore"
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

	filter := moncore.Filter_MatchAll()

	filter.Add("unwanted_field1", moncore.Filterlet_new().Exists(false))

	filter.Add("name",
		moncore.Filterlet_new().
			Exists(true).
			NotEquals("Thanos").
			RegexMatches("^T.*"),
	)

	Documents := Moncore.Database("users").Collection("accounts").Query(filter)

	C.WriteString("<h4> Documents in users/accounts without a field named 'unwanted_field1' and a field named 'name' exists which starts with 'T' but not equal to 'Thanos'  </h4> ")

	C.WriteString(` <pre> 
	filter := moncore.Filter_MatchAll()

	filter.Add("unwanted_field1", moncore.Filterlet_new().Exists(false))

	filter.Add("name",
		moncore.Filterlet_new().
			Exists(true).
			NotEquals("Thanos").
			RegexMatches("^T.*"),
	)

	Documents := Moncore.Database("users").Collection("accounts").Query(filter)
	</pre> `)

	C.WriteString("<h4> Output 'Documents' : </h4>")

	C.WriteString("<pre>")
	C.WriteJSONBeautified(Documents)
	C.WriteString("</pre>")

	C.HTMLEnd()

}

func API_List_Collections(C *APICall) {

	if len(C.Params) == 1 {
		C.WriteJSONBeautified(Moncore.Database(C.Params[0]).ListCollectionNames())

	} else {
		C.WriteError("Bad Request", errors.New("endpoints : mini/ls/<db>"), 400)
		return
	}

}

func API_List_Documents(C *APICall) {

	Q := map[string][]string(C.HTTPRequest.URL.Query())

	C.WriteJSONBeautified(Q)

	F := moncore.Filter_FromQueryStrings(Q)

	C.WriteString("\n\n")

	if len(C.Params) == 2 {
		Docs := Moncore.Database(C.Params[0]).Collection(C.Params[1]).Query(F)
		C.WriteJSONBeautified(Docs)

	} else {
		C.WriteError("Bad Request", errors.New("endpoint : mini/ls/<db>/<collection>"), 400)
		return
	}

}

func API_Set_Document(C *APICall) {

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

	C.WriteJSONBeautified(NewDocKey)

}
