package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

var _ = net.Listen
var _ = os.Exit

func main() {

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	} else {
		request := make([]byte, 1024)
		var response []byte
		// var response []byte
		n, err := conn.Read(request)
		if err != nil {
			fmt.Println("error reading message")
			os.Exit(1)
		}
		fmt.Println(string(request[:n]))
		if strings.Contains(string(request[:n]), "abcdefg") {

			response = []byte("HTTP/1.1 404 Not Found\r\n\r\n")
		} else {
			response = []byte("HTTP/1.1 200 OK\r\n\r\n")

		}
		_, err = conn.Write(response)
		if err != nil {
			fmt.Println("Couldn't respond")
		}
	}
}
