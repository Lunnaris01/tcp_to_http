package request

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"log"
	"slices"
	"strconv"
	"strings"
)

type Request struct {
	RequestLine   RequestLine
	Headers       headers.Headers
	Body          []byte
	RequestStatus int // 0 : reading RequestLine, 1: Reading Headers, 2: Reading Body, 3: done
}

func (r *Request) parse(data []byte) (int, error) {
	if loc := bytes.Index(data, []byte("\r\n")); loc != -1 {

		requestLine, err := ParseRequestLine(string(data[:loc]))
		if err != nil {
			return 0, fmt.Errorf("Failed to Parse Request line with err %v", err)

		}
		r.RequestLine = *requestLine
		r.RequestStatus = 1
		return loc + 2, nil
	}

	return 0, nil
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {

	request := Request{}
	request.Headers = headers.NewHeaders()
	buffer := make([]byte, 1024)
	readToIndex := 0
	parsedBytes := 0
	var n int
	var err error
	for request.RequestStatus < 3 {
		if readToIndex >= len(buffer) {
			newBuffer := make([]byte, len(buffer)*2) // double the size
			copy(newBuffer, buffer[:readToIndex])    // copy only valid data
			buffer = newBuffer
		}
		log.Printf("Attempting to read from connection, current readToIndex = %d if parsedBytes: %d == 0", readToIndex, parsedBytes)
		n = 0
		if parsedBytes == 0 {
			n, err = reader.Read(buffer[readToIndex:])
		}
		readToIndex += n

		if err != nil {
			if err == io.EOF {
				if request.RequestStatus == 1 {
					_, err = ParseContent(&request, buffer[:readToIndex]) // Try to parse twice for no header+no body requests..
					if err != nil {
						return &Request{}, fmt.Errorf("Failed to parse Content, maybe parts of the request are missing")
					}
				}
				_, err = ParseContent(&request, buffer[:readToIndex])
				if err != nil || request.RequestStatus != 3 {
					return &Request{}, fmt.Errorf("Failed to parse Content, maybe parts of the request are missing")
				}
				break
			}
			log.Fatalf("Failed to read: %v", err)
		}
		log.Printf("Running Parse with request: %d, for buffer %s", request.RequestStatus, buffer[:readToIndex])
		parsedBytes, err = ParseContent(&request, buffer[:readToIndex])
		log.Printf("Parsed %d bytes\n", parsedBytes)
		if err != nil {
			return &Request{}, fmt.Errorf("Failed to parse Content")
		}
		if parsedBytes > 0 {
			copy(buffer, buffer[parsedBytes:])
			readToIndex -= parsedBytes
		}
	}

	return &request, nil

}

func ParseContent(r *Request, buffer []byte) (int, error) {
	log.Printf("ParseContent called: RequestStatus=%d, Buffer='%s'", r.RequestStatus, string(buffer))
	if r.RequestStatus == 0 {
		return r.parse(buffer)
	}
	if r.RequestStatus == 1 {
		n, done, err := r.Headers.Parse(buffer)
		log.Printf("Read N: %d bytes with err: %v and done: %v", n, err, done)
		if done {
			r.RequestStatus = 2
		}
		return n, err

	}
	if r.RequestStatus == 2 {
		return parseBody(r, buffer)
	}

	return 0, fmt.Errorf("Unknown Status")
}

func ParseRequestLine(reqLineString string) (*RequestLine, error) {
	lineContent := strings.Split(reqLineString, " ")
	fmt.Println(lineContent)
	if len(lineContent) != 3 {
		return &RequestLine{}, fmt.Errorf("Malformated request line with len %d", len(reqLineString))
	}
	ret_req_line := &RequestLine{
		HttpVersion:   lineContent[2],
		RequestTarget: lineContent[1],
		Method:        lineContent[0],
	}

	valid_methods := []string{"POST", "PUT", "GET", "UPDATE"}
	if !(slices.Contains(valid_methods, ret_req_line.Method)) {
		return &RequestLine{}, fmt.Errorf("Invalid Method detected")
	}
	if ret_req_line.HttpVersion != "HTTP/1.1" {
		return &RequestLine{}, fmt.Errorf("Malformated HttpVersion detected")
	} else {
		ret_req_line.HttpVersion = "1.1"
	}
	if !(strings.ToUpper(ret_req_line.Method) == ret_req_line.Method) {
		return &RequestLine{}, fmt.Errorf("Malformated Method line detected")
	}
	return ret_req_line, nil
}

func parseBody(r *Request, buffer []byte) (int, error) {
	log.Print("Parsing Body!")
	cLen, ok := r.Headers.Get("Content-Length")
	if ok != nil {
		log.Print("No Content-Length Header found! Assuming Empty Body!")
		r.RequestStatus = 3
		return 0, nil
	}
	expected_length, err := strconv.Atoi(cLen)
	if err != nil {
		return 0, fmt.Errorf("Malformed content of Content-Length Header")
	}
	if len(buffer) == expected_length {
		log.Print("Content length matching Content-Length header")
		r.Body = make([]byte, len(buffer))
		copy(r.Body, buffer)
		r.RequestStatus = 3
		return expected_length, nil
	}
	if len(buffer) > expected_length {
		return 0, fmt.Errorf("Body longer than expected")
	}
	return 0, nil
}
