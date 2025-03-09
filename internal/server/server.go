package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

const returnString = "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!"

type Server struct {
	port     int
	listener net.Listener
	running  bool
}

func Serve(port int) (*Server, error) {
	portStr := strconv.Itoa(port)

	listener, err := net.Listen("tcp", ":"+portStr)
	if err != nil {
		return &Server{}, fmt.Errorf("Failed to start server: %v", err)
	}
	server := Server{port, listener, true}
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
		reader, err := s.listener.Accept()
		if err != nil {
			log.Fatal("failed to accept connection")
		}
		s.handle(reader)
		reader.Close()

		fmt.Println("Connection accepted successfully")
		//req, err := request.RequestFromReader(reader)
		//reader.Close()
		if err != nil {
			log.Fatalf("Failed to read or parse request: %v", err)
		}
		//rLine := req.RequestLine
		fmt.Println("parsed content successfully")

		//fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", rLine.Method, rLine.RequestTarget, rLine.HttpVersion)
		//fmt.Printf("Headers:\n")
		//for key, value := range req.Headers {
		//	fmt.Printf("- %s: %s\n", key, value)
		//}
		//fmt.Printf("Body:\n%s\n", req.Body)

	}
}

func (s *Server) handle(conn net.Conn) {
	conn.Write([]byte(returnString))
}
