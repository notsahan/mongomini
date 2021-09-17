package endpoints

// func api_POST_Auth(C *API_Call_Receieved) {
// 	return StringResponse("API POST recieved for api/auth/" + C + " with " + strconv.Itoa(len(body)) + " bytes of body")
// }

func API_Auth(C *APICall) {
	C.WriteString("API GET recieved. Path : " + C.Path)
}
