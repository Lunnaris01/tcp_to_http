package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("failed to create listener")
	}
	defer listener.Close()
	for {
		reader, err := listener.Accept()
		if err != nil {
			log.Fatal("failed to accept connection")
		}
		fmt.Println("Connection accepted successfully")
		req, err := request.RequestFromReader(reader)
		if err != nil {
			log.Fatalf("Failed to read or parse request: %v", err)
		}
		rLine := req.RequestLine

		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", rLine.Method, rLine.RequestTarget, rLine.HttpVersion)
		fmt.Printf("Headers:\n")
		for key, value := range req.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}
		fmt.Printf("Body:\n%s\n", req.Body)

	}

}
