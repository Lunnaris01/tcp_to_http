package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"log"
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

/*
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
*/
func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	contentLenStr := strconv.Itoa(contentLen)
	headers.AddKey("Content-Length", contentLenStr)
	headers.AddKey("Content-Type", "text/plain")
	headers.AddKey("Connection", "close")
	return headers
}

func GetDefaultChunkedHeaders() headers.Headers {
	headers := headers.NewHeaders()
	headers.AddKey("Transfer-Encoding", "chunked")
	headers.AddKey("Content-Type", "text/plain")
	headers.AddKey("Connection", "close")
	return headers
}

/*
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
*/

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
	writerStateDone
	writerStatusTrailer
)

type Writer struct {
	Connection       io.Writer
	writerState      writerState
	pos              int
	numBytesPerWrite int
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		Connection:       w,
		writerState:      writerStateStatusLine,
		pos:              0,
		numBytesPerWrite: 1024,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != writerStateStatusLine {
		return fmt.Errorf("Unable to write StatusLine while in state %v", w.writerState)
	}
	defer func() { w.writerState = writerStateHeaders }()
	val, ok := mapStatusCode[statusCode]
	var statusLine []byte
	var err error
	if ok {
		statusLine = []byte("HTTP/1.1 " + val + "\r\n")
	} else {
		statusLine = []byte("HTTP/1.1 " + strconv.Itoa(int(statusCode)) + "\r\n")
	}

	fmt.Printf("Writing status Line: %s", string(statusLine))
	_, err = w.Connection.Write(statusLine)
	if err != nil {
		fmt.Println("Error writing status line:", err)
	}
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != writerStateHeaders {
		log.Printf("Unable to write Headers while in state %v", w.writerState)
		return fmt.Errorf("Unable to write Headers while in state %v", w.writerState)
	}
	defer func() { w.writerState = writerStateBody }()

	for header, val := range headers {
		_, err := w.Connection.Write([]byte(header + ": " + val + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err := w.Connection.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	return nil

}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("Unable to write Body while in state %v", w.writerState)
	}
	defer func() { w.writerState = writerStateDone }()
	return w.Connection.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (n int, err error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("Unable to write Body while in state %v", w.writerState)
	}

	// Write the chunk size in hexadecimal format
	sizeHex := fmt.Sprintf("%x\r\n", len(p))
	_, err = w.Connection.Write([]byte(sizeHex))
	if err != nil {
		return 0, fmt.Errorf("error writing chunk size: %v", err)
	}

	n, err = w.Connection.Write(p)
	if err != nil {
		return 0, fmt.Errorf("error %v", err)
	}

	// Write the trailing CRLF
	_, err = w.Connection.Write([]byte("\r\n"))
	if err != nil {
		return 0, fmt.Errorf("error writing chunk trailer: %v", err)
	}

	return len(p), nil

}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	n, err := w.Connection.Write([]byte("0\r\n"))
	if err != nil {
		return n, err
	}
	w.writerState = writerStatusTrailer
	return n, nil

}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.writerState != writerStatusTrailer {
		return fmt.Errorf("Unable to write Trailers while in state %v", w.writerState)
	}
	w.writerState = writerStateHeaders
	err := w.WriteHeaders(h)
	if err != nil {
		return err
	}
	w.writerState = writerStateDone
	_, err = w.Connection.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	return nil
}
