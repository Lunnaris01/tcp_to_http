package server

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log"
	"net"
	"strconv"
)

const returnString = "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!"

type Server struct {
	port     int
	listener net.Listener
	handler  Handler
	running  bool
}

func Serve(handler Handler, port int) (*Server, error) {
	portStr := strconv.Itoa(port)

	listener, err := net.Listen("tcp", ":"+portStr)
	if err != nil {
		return &Server{}, fmt.Errorf("Failed to start server: %v", err)
	}
	server := Server{port, listener, handler, true}
	go server.listen()
	return &server, nil
}

func (s *Server) Close() error {
	if !s.running {
		return fmt.Errorf("Server already closed!")
	}
	fmt.Println("Closing Server")
	s.running = false
	return s.listener.Close()
}

func (s *Server) listen() {
	for s.running {
		connection, err := s.listener.Accept()
		if err != nil {
			log.Fatal("failed to accept connection")
		}

		fmt.Println("Connection accepted successfully")
		s.handle(connection)
		connection.Close()
		if err != nil {
			log.Fatalf("Failed to read or parse request: %v", err)
		}
		//rLine := req.RequestLine

	}
}

func (s *Server) handle(connection net.Conn) {
	req, err := request.RequestFromReader(connection)

	log.Println("parsed content successfully")
	rline := req.RequestLine
	log.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", rline.Method, rline.RequestTarget, rline.HttpVersion)
	log.Printf("Headers:\n")
	for key, value := range req.Headers {
		log.Printf("- %s: %s\n", key, value)
	}
	log.Printf("Body:\n%s\n", req.Body)

	if err != nil {
		hErr := &HandlerError{
			ErrorCode:    response.StatusBadRequest,
			ErrorMessage: err.Error(),
		}
		log.Print("First hErr")
		hErr.Write(connection)
	}
	buf := bytes.NewBuffer([]byte("All good, frfr\n"))
	hErr := s.handler(buf, req)
	if hErr != nil {
		log.Print("Second hERR")
		hErr.Write(connection)
		return
	}

	b := buf.Bytes()

	response.WriteStatusLine(connection, response.StatusOk)
	headers := response.GetDefaultHeaders(len(b))
	response.WriteHeaders(connection, headers)
	connection.Write(b)
	return
}

type HandlerError struct {
	ErrorCode    response.StatusCode
	ErrorMessage string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func DefaultHandler(w io.Writer, req *request.Request) *HandlerError {
	handlerError := HandlerError{}
	if req.RequestLine.Method == "GET" {
		if req.RequestLine.RequestTarget == "/yourproblem" {
			handlerError.ErrorCode = 400
			handlerError.ErrorMessage = "Your problem is not my problem\n"
			return &handlerError
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			handlerError.ErrorCode = 500
			handlerError.ErrorMessage = "Woopsie, my bad\n"
			return &handlerError
		}
	}
	return nil

}

func (handlerError HandlerError) Write(w io.Writer) {
	log.Printf("Writing with handlerError:")
	response.WriteStatusLine(w, handlerError.ErrorCode)
	heMessageBytes := []byte(handlerError.ErrorMessage)
	response.WriteHeaders(w, response.GetDefaultHeaders(len(heMessageBytes)))
	w.Write(heMessageBytes)
}
