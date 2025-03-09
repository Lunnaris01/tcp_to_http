package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type StatusCode int

const (
	StatusOk                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusNotFound            StatusCode = 404
	StatusInternalServerError StatusCode = 500
)

var mapStatusCode = map[StatusCode]string{
	StatusOk:                  "200 OK",
	StatusBadRequest:          "400 Bad Request",
	StatusNotFound:            "404 Not Found",
	StatusInternalServerError: "500 Internal Server Error",
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) {
	val, ok := mapStatusCode[statusCode]
	var statusLine []byte
	var err error
	if ok {
		statusLine = []byte("HTTP/1.1 " + val + "\r\n")
	} else {
		statusLine = []byte("HTTP/1.1 " + strconv.Itoa(int(statusCode)) + "\r\n")
	}

	fmt.Printf("Writing status Line: %s", string(statusLine))
	_, err = w.Write(statusLine)
	if err != nil {
		fmt.Println("Error writing status line:", err)
	}

}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	contentLenStr := strconv.Itoa(contentLen)
	headers["Content-Length"] = contentLenStr
	headers["Content-Type"] = "text/plain"
	headers["Connection"] = "close"
	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for header, val := range headers {
		_, err := w.Write([]byte(header + ": " + val + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	return nil
}
