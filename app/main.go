package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

var _ = net.Listen
var _ = os.Exit

func main() {

	for {
		l, err := net.Listen("tcp", "127.0.0.1:4221")
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		} else {
			go handleRequest(conn)

		}
	}

}
func handleRequest(conn net.Conn) {
	request := make([]byte, 1024)
	var response []byte
	n, err := conn.Read(request)
	if err != nil {
		fmt.Println("error reading message")
		os.Exit(1)
	}
	requestTarget := getRequestTarget(request[:n])

	if string(requestTarget) == "/" {
		response = []byte("HTTP/1.1 200 OK\r\n\r\n")

	} else if strings.HasPrefix(string(requestTarget), "/echo/") {
		str := strings.Split(string(requestTarget), "/")

		response = []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length:" + strconv.Itoa(len(str[2])) + "\r\n\r\n" + str[2])
	} else if strings.HasPrefix(string(requestTarget), "/user-agent") {
		response = handleUserAgent(request, n)
	} else {
		response = []byte("HTTP/1.1 404 Not Found\r\n\r\n")

	}
	_, err = conn.Write(response)
	if err != nil {
		fmt.Println("Couldn't respond")
	}
}
func getRequestTarget(request []byte) []byte {
	var requestTarget []byte

	startCapturingTarget := false
	for _, ch := range request {
		if ch == '/' && !startCapturingTarget {
			startCapturingTarget = true
			requestTarget = append(requestTarget, ch)
		} else if ch != ' ' && startCapturingTarget {
			requestTarget = append(requestTarget, ch)
		} else if ch == ' ' && startCapturingTarget {
			return requestTarget
		}
	}
	return requestTarget
}
func handleUserAgent(request []byte, n int) []byte {
	var response []byte
	headers := strings.Split(string(request[:n]), "Host: localhost:4221\r\n")[1]

	userAgentIndex := strings.Index(headers, "User-Agent")
	userAgent := headers[userAgentIndex:]
	userAgent = userAgent[12:strings.Index(userAgent, "\r\n")]

	response = []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + strconv.Itoa(len(userAgent)) + "\r\n\r\n" + userAgent)
	return response
}
