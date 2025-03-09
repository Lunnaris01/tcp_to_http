package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
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
	w := response.NewWriter(connection)
	if err != nil {
		w.WriteStatusLine(response.StatusBadRequest)
		body := []byte(fmt.Sprintf("Error parsing request: %v", err))
		w.WriteHeaders(response.GetDefaultHeaders(len(body)))
		w.WriteBody(body)
		return

	}

	s.handler(w, req)

	return
}

type Handler func(w *response.Writer, req *request.Request)

/*
func DefaultHandler(w response.Writer, req *request.Request) *HandlerError {
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
*/
