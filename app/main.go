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
	l, err := net.Listen("tcp", "127.0.0.1:4221")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	for {

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
	requestType := getRequestType(request[:n])
	requestTarget := getRequestTarget(request[:n])
	if requestType == "GET" {
		if string(requestTarget) == "/" {
			response = []byte("HTTP/1.1 200 OK\r\n\r\n")

		} else if strings.HasPrefix(string(requestTarget), "/echo/") {
			str := strings.Split(string(requestTarget), "/")

			response = []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length:" + strconv.Itoa(len(str[2])) + "\r\n\r\n" + str[2])
		} else if strings.HasPrefix(string(requestTarget), "/user-agent") {
			response = handleUserAgent(request, n)
		} else if strings.HasPrefix(string(requestTarget), "/files") {
			targetFile := "/tmp/data/codecrafters.io/http-server-tester/" + strings.Split(string(requestTarget), "/")[2]
			file, err := os.OpenFile(targetFile, os.O_RDONLY, 0777)
			if err != nil {
				response = []byte("HTTP/1.1 404 Not Found\r\n\r\n")
			} else {
				fileContent := make([]byte, 1024)
				nChars, _ := file.Read(fileContent)
				response = []byte("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: " + strconv.Itoa(nChars) + "\r\n\r\n" + string(fileContent[:nChars]))
			}
		} else {
			response = []byte("HTTP/1.1 404 Not Found\r\n\r\n")

		}
	} else if requestType == "POST" {
		if strings.HasPrefix(string(requestTarget), "/files") {
			targetFile := "/tmp/data/codecrafters.io/http-server-tester/" + strings.Split(string(requestTarget), "/")[2]
			file, err := os.OpenFile(targetFile, os.O_CREATE, 0777)
			if err != nil {
				response = []byte("HTTP/1.1 404 Not Found\r\n\r\n")
			} else {
				response = writePOSTToFile(file, request)
			}
		}

		_, err = conn.Write(response)
		if err != nil {
			fmt.Println("Couldn't respond")
		}
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
func getRequestType(request []byte) string {
	var requestType string
	for _, ch := range request {
		if ch == ' ' {
			break

		} else {
			requestType = requestType + string(ch)
		}
	}
	return requestType
}
func writePOSTToFile(file *os.File, request []byte) []byte {
	var fileContent []byte

	//get content length from request
	//get data from request + content length
	//write data to file
	//return response
	var response []byte
	contentLengthHeaderBytes := string(request[strings.Index(string(request), "Content-Length")+16 : strings.Index(string(request), "Content-Length")+19])
	contentLengthHeaderBytes = strings.Trim(contentLengthHeaderBytes, "\r")
	contentLengthHeaderBytes = strings.Trim(contentLengthHeaderBytes, " ")

	contentLengthHeader, _ := strconv.Atoi(contentLengthHeaderBytes)
	fileContent = request[len(request)-contentLengthHeader-1:]
	file, _ = os.OpenFile(file.Name(), os.O_RDWR, 0777)
	_, err := file.Write(fileContent)

	if err != nil {
		response = []byte("HTTP/1.1 404 Not Found\r\n\r\n")

	} else {
		response = []byte("HTTP/1.1 201 Created\r\n\r\n")

	}
	return response
}

// func getData(request []byte, contentLength int) string {
// 	var data string
// 	data = string(request[len(request)-contentLength-1:])
// 	return data
// }
