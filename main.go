package main

import (
	"errors"
	"log"
	handler "mongomini/api"
	"net"
	"net/http"
)

var HTTPPort string = "49525"

func main() {

	FullMux := http.NewServeMux()

	FullMux.HandleFunc("/", handler.ServeRequest)

	handler.InitAll()

	println("API running on http://localhost:" + HTTPPort)

	if err := http.ListenAndServe(":"+HTTPPort, FullMux); err != nil {
		log.Fatal(err)
	}

}

// Get preferred outbound ip of this machine
func GetOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		println("Cannot get my IP address. May be offline")
		// PrintError(err)
		return "", errors.New("error")
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}
