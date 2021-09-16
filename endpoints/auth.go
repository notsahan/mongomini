package endpoints

import "strconv"

func api_POST_Auth(path string, body []byte) Response {
	return StringResponse("API POST recieved for api/auth/" + path + " with " + strconv.Itoa(len(body)) + " bytes of body")
}

func api_GET_Auth(path string) Response {
	return StringResponse("API GET recieved. Path : api/auth/" + path)
}
