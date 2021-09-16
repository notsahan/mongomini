package endpoints

func api_GET_Hello(path string) Response {
	return StringResponse("api_GET_Hello recieved. Path : " + path)
}
