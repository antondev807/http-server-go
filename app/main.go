package main

import (
	"bytes"
	"compress/gzip"
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

	for {
		n, err := conn.Read(request)

		if err == nil {
			requestHeaders := getRequestHeaders(request[:n])

			requestType := getRequestType(request[:n])

			requestTarget := getRequestTarget(request[:n])
			if requestType == "GET" {
				if string(requestTarget) == "/" {
					if requestHeaders["Connection"] == "close" {
						response = []byte("HTTP/1.1 200 OK\r\n\r\nConnection: close")

						break
					} else {
						response = []byte("HTTP/1.1 200 OK\r\n\r\n")

					}

				} else if strings.HasPrefix(string(requestTarget), "/echo/") {
					// fmt.Println(requestHeaders["Content-Encoding"])
					str := strings.Split(string(requestTarget), "/")
					if strings.Contains(requestHeaders["Accept-Encoding"], "gzip") {
						// fmt.Println("here")
						var buf bytes.Buffer

						writer := gzip.NewWriter(&buf)
						_, err := writer.Write([]byte(str[2]))
						if err != nil {
							log.Fatal(err)
						}

						_ = writer.Close()
						compressed := buf.Bytes()
						if requestHeaders["Connection"] == "close" {
							response = []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nConnection: close\r\nContent-Encoding: gzip\r\nContent-Length:" + strconv.Itoa(len(compressed)) + "\r\n\r\n" + string(compressed))

							break
						} else {
							response = []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Encoding: gzip\r\nContent-Length:" + strconv.Itoa(len(compressed)) + "\r\n\r\n" + string(compressed))

						}
					} else {
						if requestHeaders["Connection"] == "close" {
							response = []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length:" + strconv.Itoa(len(str[2])) + "\r\nConnection: close\r\n\r\n" + str[2])

							break
						} else {
							response = []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length:" + strconv.Itoa(len(str[2])) + "\r\n\r\n" + str[2])
						}
					}

				} else if strings.HasPrefix(string(requestTarget), "/user-agent") {
					response,shouldClose := handleUserAgent(request, n)
					if shouldClose{
						break
					}
				} else if strings.HasPrefix(string(requestTarget), "/files") {
					targetFile := "/tmp/data/codecrafters.io/http-server-tester/" + strings.Split(string(requestTarget), "/")[2]
					file, err := os.OpenFile(targetFile, os.O_RDONLY, 0777)

					if err != nil {
						response = []byte("HTTP/1.1 404 Not Found\r\n\r\n")
					} else {

						fileContent := make([]byte, 1024)
						nChars, _ := file.Read(fileContent)
						file.Close()
						response = []byte("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: " + strconv.Itoa(nChars) + "\r\n\r\n" + string(fileContent[:nChars]))
					}
				} else {
					response = []byte("HTTP/1.1 404 Not Found\r\n\r\n")

				}
			} else if requestType == "POST" {
				if strings.HasPrefix(string(requestTarget), "/files") {
					targetFile := "/tmp/data/codecrafters.io/http-server-tester/" + strings.Split(string(requestTarget), "/")[2]
					file, err := os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)

					if err != nil {
						response = []byte("HTTP/1.1 404 Not Found\r\n\r\n")
					} else {

						response = writePOSTToFile(file, request)
						file.Close()
					}
				}

			}
			_, err = conn.Write(response)
			if err != nil {
				fmt.Println("Couldn't respond")
			}

		}

	}
	conn.Close()

	// conn, err = l.Accept()
	// if err != nil {
	// 	fmt.Println("Error accepting connection: ", err.Error())
	// 	os.Exit(1)
	// } else {
	// 	go handleRequest(conn, l)

	// }

}

func getRequestHeaders(request []byte) map[string]string {
	requestHeaders := make(map[string]string)
	requestString := string(request)
	// fmt.Println(requestString)

	requestLines := strings.Split(requestString, "\r\n")
	for _, line := range requestLines {

		if strings.Index(line, ":") != -1 {
			parts := strings.SplitN(line, ":", 2)

			key := parts[0]
			value := strings.TrimSpace(parts[1])

			requestHeaders[key] = value
		}

	}
	// requestHeaders["test"] = "foo"
	// fmt.Println(requestHeaders)
	return requestHeaders

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
func handleUserAgent(request []byte, n int) []byte ,bool{
	var response []byte
	headers := strings.Split(string(request[:n]), "Host: localhost:4221\r\n")[1]

	userAgentIndex := strings.Index(headers, "User-Agent")
	userAgent := headers[userAgentIndex:]
	userAgent = userAgent[12:strings.Index(userAgent, "\r\n")]
	if requestHeaders["Connection"] == "close" {
		shouldClose:=true
		response = []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nConnection: close\r\nContent-Length: " + strconv.Itoa(len(userAgent)) + "\r\n\r\n" + userAgent)

	} else {
		shouldClose:=false
		response = []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + strconv.Itoa(len(userAgent)) + "\r\n\r\n" + userAgent)

	}
	return response, shouldClose
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
	trimmedRequest := strings.Trim(string(request), string('\x00'))
	request = []byte(trimmedRequest) //this is probably buggy as it assumes the byte slice is initialized to this char always
	//get content length from request
	//get data from request + content length
	//write data to file
	//return response
	var response []byte
	contentLengthHeaderBytes := string(request[strings.Index(string(request), "Content-Length")+16 : strings.Index(string(request), "Content-Length")+19])
	contentLengthHeaderBytes = strings.Trim(contentLengthHeaderBytes, "\r")
	contentLengthHeaderBytes = strings.Trim(contentLengthHeaderBytes, " ")

	contentLengthHeader, _ := strconv.Atoi(contentLengthHeaderBytes)
	fileContent = request[len(request)-contentLengthHeader:]
	file, _ = os.OpenFile(file.Name(), os.O_RDWR, 0777)
	_, err := file.Write(fileContent)

	if err != nil {
		response = []byte("HTTP/1.1 404 Not Found\r\n\r\n")

	} else {
		defer file.Close()

		response = []byte("HTTP/1.1 201 Created\r\n\r\n")

	}
	return response
}

// func getData(request []byte, contentLength int) string {
// 	var data string
// 	data = string(request[len(request)-contentLength-1:])
// 	return data
// }
