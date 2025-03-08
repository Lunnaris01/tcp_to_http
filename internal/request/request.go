package request

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"slices"
	"strings"
)

type Request struct {
	RequestLine   RequestLine
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
	buffer := make([]byte, 8)
	readToIndex := 0
	for request.RequestStatus == 0 {
		if readToIndex >= len(buffer) {
			newBuffer := make([]byte, len(buffer)*2) // double the size
			copy(newBuffer, buffer[:readToIndex])    // copy only valid data
			buffer = newBuffer
		}
		n, err := reader.Read(buffer[readToIndex:])
		readToIndex += n
		fmt.Printf("Bufferlen: %v, Buffercap: %v\n", len(buffer), cap(buffer))
		fmt.Println(string(buffer))
		if err != nil {
			if err == io.EOF {
				_, err := request.parse(buffer[:readToIndex])
				if err != nil {
					fmt.Println("F")
					return &Request{}, fmt.Errorf("Failed to parse after hitting EOF")
				}
				break
			}
			log.Fatalf("Failed to read: %v", err)
		}

		parsedBytes, err := request.parse(buffer)
		//fmt.Printf("Processed N bytes: %d\n", n)
		//fmt.Printf("%v\n", strings.Contains(string(buffer), "\r\n"))
		//fmt.Printf("%v\n", bytes.Contains(buffer, []byte("\r\n")))
		//fmt.Printf("Requeststatus: %v\n", request.RequestStatus)
		if err != nil {
			log.Fatalf("Failed to parse Requestline %v", err)
		}
		if parsedBytes > 0 {
			copy(buffer, buffer[parsedBytes:readToIndex])
			readToIndex -= parsedBytes
		}
	}

	return &request, nil

}

func ParseRequestLine(reqLineString string) (*RequestLine, error) {
	fmt.Println((reqLineString))
	lineContent := strings.Split(reqLineString, " ")
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
	fmt.Printf("HttpVersion: %s\nExpected: %s\nEquality Check: %v\n", ret_req_line.HttpVersion, "HTTP/1.1", ret_req_line.HttpVersion == "HTTP/1.1")
	fmt.Printf("Len1: %d, Len2: %d\n", len(ret_req_line.HttpVersion), len("HTTP/1.1"))
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
