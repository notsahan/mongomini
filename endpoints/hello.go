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

	F := moncore.Filter_MatchAll()

	Q := C.HTTPRequest.URL.Query()

	C.WriteJSONBeautified(Q)

	for qk, qvs := range Q {
		if len(qk) == 0 {
			continue
		}

		fl := moncore.Filterlet_new()
		if len(qvs) == 0 { // Exist
			fl.Exists(true)

		} else {
			for _, v := range qvs {
				if strings.HasPrefix(v, "=") {
					fl.Equals(v[1:])

				} else if strings.HasPrefix(v, "-") {
					fl.NotEquals(v[1:])

				} else if v == "exist" || len(v) == 0 {
					fl.Exists(true)

				} else if v == "not-exist" {
					fl.Exists(false)

				}

			}
		}

		if len(fl.Querylet) != 0 {
			F.Add(qk, fl)
		}

	}

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
